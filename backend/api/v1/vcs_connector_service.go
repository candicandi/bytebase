package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/common/log"
	"github.com/bytebase/bytebase/backend/plugin/vcs"
	"github.com/bytebase/bytebase/backend/plugin/vcs/azure"
	"github.com/bytebase/bytebase/backend/plugin/vcs/bitbucket"
	"github.com/bytebase/bytebase/backend/plugin/vcs/github"
	"github.com/bytebase/bytebase/backend/plugin/vcs/gitlab"
	"github.com/bytebase/bytebase/backend/store"
	storepb "github.com/bytebase/bytebase/proto/generated-go/store"
	v1pb "github.com/bytebase/bytebase/proto/generated-go/v1"
)

// VCSConnectorService implements the vcs connector service.
type VCSConnectorService struct {
	v1pb.UnimplementedVCSConnectorServiceServer
	store *store.Store
}

// NewVCSConnectorService creates a new VCSConnectorService.
func NewVCSConnectorService(store *store.Store) *VCSConnectorService {
	return &VCSConnectorService{
		store: store,
	}
}

// CreateVCSConnector creates a vcs connector.
func (s *VCSConnectorService) CreateVCSConnector(ctx context.Context, request *v1pb.CreateVCSConnectorRequest) (*v1pb.VCSConnector, error) {
	if request.VcsConnector == nil {
		return nil, status.Errorf(codes.InvalidArgument, "vcs connector must be set")
	}

	setting, err := s.store.GetWorkspaceGeneralSetting(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find workspace setting: %v", err)
	}
	if setting.ExternalUrl == "" {
		return nil, status.Errorf(codes.FailedPrecondition, setupExternalURLError)
	}

	projectResourceID, err := common.GetProjectID(request.Parent)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	project, err := s.store.GetProjectV2(ctx, &store.FindProjectMessage{
		ResourceID: &projectResourceID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get project with resource id %q, err: %v", projectResourceID, err)
	}
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project with resource id %q not found", projectResourceID)
	}
	if project.Deleted {
		return nil, status.Errorf(codes.NotFound, "project with resource id %q had deleted", projectResourceID)
	}

	vcsConnector, err := s.store.GetVCSConnector(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID, ResourceID: &request.VcsConnectorId})
	if err != nil {
		return nil, err
	}
	if vcsConnector != nil {
		return nil, status.Errorf(codes.AlreadyExists, "vcs connector %q already exists", request.VcsConnectorId)
	}

	vcsResourceID, err := common.GetVCSProviderID(request.GetVcsConnector().GetVcsProvider())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	vcsProvider, err := s.store.GetVCSProvider(ctx, &store.FindVCSProviderMessage{ResourceID: &vcsResourceID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find vcs: %s", err.Error())
	}
	if vcsProvider == nil {
		return nil, status.Errorf(codes.NotFound, "vcs %s not found", vcsResourceID)
	}

	// Check branch existence.
	if err := checkBranchExistence(
		ctx,
		vcsProvider,
		request.GetVcsConnector().ExternalId,
		request.GetVcsConnector().Branch,
	); err != nil {
		return nil, err
	}

	baseDirectory := request.GetVcsConnector().BaseDirectory
	if !strings.HasPrefix(baseDirectory, "/") {
		return nil, status.Errorf(codes.InvalidArgument, `base directory should start with "/"`)
	}
	if strings.HasSuffix(baseDirectory, "/") {
		return nil, status.Errorf(codes.InvalidArgument, `base directory should not end with "/"`)
	}

	workspaceID, err := s.store.GetWorkspaceID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find workspace id with error: %v", err.Error())
	}
	secretToken, err := common.RandomString(gitlab.SecretTokenLength)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate random secret token for vcs with error: %v", err.Error())
	}
	vcsConnectorCreate := &store.VCSConnectorMessage{
		ProjectID:     project.ResourceID,
		ResourceID:    request.VcsConnectorId,
		VCSUID:        vcsProvider.ID,
		VCSResourceID: vcsProvider.ResourceID,
		Payload: &storepb.VCSConnector{
			Title:              request.GetVcsConnector().Title,
			FullPath:           request.GetVcsConnector().FullPath,
			WebUrl:             request.GetVcsConnector().WebUrl,
			Branch:             request.GetVcsConnector().Branch,
			BaseDirectory:      request.GetVcsConnector().BaseDirectory,
			ExternalId:         request.GetVcsConnector().ExternalId,
			WebhookSecretToken: secretToken,
			DatabaseGroup:      request.GetVcsConnector().DatabaseGroup,
		},
	}

	// Create the webhook.
	bytebaseEndpointURL := setting.GitopsWebhookUrl
	if bytebaseEndpointURL == "" {
		bytebaseEndpointURL = setting.ExternalUrl
	}
	webhookEndpointID := fmt.Sprintf("workspaces/%s/projects/%s/vcsConnectors/%s", workspaceID, project.ResourceID, request.VcsConnectorId)
	webhookID, err := createVCSWebhook(
		ctx,
		vcsProvider,
		webhookEndpointID,
		secretToken,
		vcsConnectorCreate.Payload.ExternalId,
		bytebaseEndpointURL,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create webhook for project %s with error: %v", vcsConnectorCreate.ProjectID, err.Error())
	}
	vcsConnectorCreate.Payload.ExternalWebhookId = webhookID

	vcsConnector, err = s.store.CreateVCSConnector(ctx, vcsConnectorCreate)
	if err != nil {
		return nil, err
	}
	v1VCSConnector, err := convertStoreVCSConnector(vcsConnector)
	if err != nil {
		return nil, err
	}
	return v1VCSConnector, nil
}

// GetVCSConnector gets a vcs connector.
func (s *VCSConnectorService) GetVCSConnector(ctx context.Context, request *v1pb.GetVCSConnectorRequest) (*v1pb.VCSConnector, error) {
	projectID, vcsConnectorID, err := common.GetProjectVCSConnectorID(request.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	project, err := s.store.GetProjectV2(ctx, &store.FindProjectMessage{
		ResourceID: &projectID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found", projectID)
	}

	vcsConnector, err := s.store.GetVCSConnector(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID, ResourceID: &vcsConnectorID})
	if err != nil {
		return nil, err
	}
	if vcsConnector == nil {
		return nil, status.Errorf(codes.NotFound, "vcs connector %q not found", vcsConnectorID)
	}
	v1VCSConnector, err := convertStoreVCSConnector(vcsConnector)
	if err != nil {
		return nil, err
	}
	return v1VCSConnector, nil
}

// GetVCSConnector gets a vcs connector.
func (s *VCSConnectorService) ListVCSConnectors(ctx context.Context, request *v1pb.ListVCSConnectorsRequest) (*v1pb.ListVCSConnectorsResponse, error) {
	projectID, err := common.GetProjectID(request.Parent)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	project, err := s.store.GetProjectV2(ctx, &store.FindProjectMessage{
		ResourceID: &projectID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found", projectID)
	}

	vcsConnectors, err := s.store.ListVCSConnectors(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID})
	if err != nil {
		return nil, err
	}

	resp := &v1pb.ListVCSConnectorsResponse{}
	for _, vcsConnector := range vcsConnectors {
		v1VCSConnector, err := convertStoreVCSConnector(vcsConnector)
		if err != nil {
			return nil, err
		}
		resp.VcsConnectors = append(resp.VcsConnectors, v1VCSConnector)
	}
	return resp, nil
}

// UpdateVCSConnector updates a vcs connector.
func (s *VCSConnectorService) UpdateVCSConnector(ctx context.Context, request *v1pb.UpdateVCSConnectorRequest) (*v1pb.VCSConnector, error) {
	if request.VcsConnector == nil {
		return nil, status.Errorf(codes.InvalidArgument, "vcs connector must be set")
	}
	if request.UpdateMask == nil {
		return nil, status.Errorf(codes.InvalidArgument, "update_mask must be set")
	}

	projectID, vcsConnectorID, err := common.GetProjectVCSConnectorID(request.GetVcsConnector().GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	project, err := s.store.GetProjectV2(ctx, &store.FindProjectMessage{
		ResourceID: &projectID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found", projectID)
	}

	vcsConnector, err := s.store.GetVCSConnector(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID, ResourceID: &vcsConnectorID})
	if err != nil {
		return nil, err
	}
	if vcsConnector == nil {
		return nil, status.Errorf(codes.NotFound, "vcs connector %q not found", vcsConnectorID)
	}

	vcsProvider, err := s.store.GetVCSProvider(ctx, &store.FindVCSProviderMessage{ResourceID: &vcsConnector.VCSResourceID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find vcs: %s", err.Error())
	}
	if vcsProvider == nil {
		return nil, status.Errorf(codes.NotFound, "vcs provider %s not found", vcsConnector.VCSResourceID)
	}

	update := &store.UpdateVCSConnectorMessage{
		ProjectID: project.ResourceID,
		UID:       vcsConnector.UID,
	}

	for _, path := range request.UpdateMask.Paths {
		switch path {
		case "branch":
			update.Branch = &request.GetVcsConnector().Branch
		case "base_directory":
			baseDir := request.GetVcsConnector().BaseDirectory
			if !strings.HasPrefix(baseDir, "/") {
				return nil, status.Errorf(codes.InvalidArgument, `base directory should start with "/"`)
			}
			update.BaseDirectory = &baseDir
		case "database_group":
			update.DatabaseGroup = &request.GetVcsConnector().DatabaseGroup
		}
	}

	// Check branch existence.
	if v := update.Branch; v != nil {
		if err := checkBranchExistence(
			ctx,
			vcsProvider,
			vcsConnector.Payload.ExternalId,
			*v,
		); err != nil {
			return nil, err
		}
	}

	if err := s.store.UpdateVCSConnector(ctx, update); err != nil {
		return nil, err
	}
	vcsConnector, err = s.store.GetVCSConnector(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID, ResourceID: &vcsConnectorID})
	if err != nil {
		return nil, err
	}

	v1VCSConnector, err := convertStoreVCSConnector(vcsConnector)
	if err != nil {
		return nil, err
	}
	return v1VCSConnector, nil
}

// DeleteVCSConnector deletes a vcs connector.
func (s *VCSConnectorService) DeleteVCSConnector(ctx context.Context, request *v1pb.DeleteVCSConnectorRequest) (*emptypb.Empty, error) {
	projectID, vcsConnectorID, err := common.GetProjectVCSConnectorID(request.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	project, err := s.store.GetProjectV2(ctx, &store.FindProjectMessage{
		ResourceID: &projectID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if project == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found", projectID)
	}

	vcsConnector, err := s.store.GetVCSConnector(ctx, &store.FindVCSConnectorMessage{ProjectID: &project.ResourceID, ResourceID: &vcsConnectorID})
	if err != nil {
		return nil, err
	}
	if vcsConnector == nil {
		return nil, status.Errorf(codes.NotFound, "vcs connector %q not found", vcsConnectorID)
	}

	vcsProvider, err := s.store.GetVCSProvider(ctx, &store.FindVCSProviderMessage{ResourceID: &vcsConnector.VCSResourceID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find vcs: %s", err.Error())
	}
	if vcsProvider == nil {
		return nil, status.Errorf(codes.NotFound, "vcs provider %d not found", vcsConnector.UID)
	}

	if err := s.store.DeleteVCSConnector(ctx, project.ResourceID, vcsConnectorID); err != nil {
		return nil, err
	}

	// Delete the webhook, and fail-open.
	webhookIDs := strings.Split(vcsConnector.Payload.ExternalWebhookId, ",")
	vcsPlugin := vcs.Get(
		vcsProvider.Type,
		vcs.ProviderConfig{InstanceURL: vcsProvider.InstanceURL, AuthToken: vcsProvider.AccessToken},
	)
	for _, webhookID := range webhookIDs {
		if err = vcsPlugin.DeleteWebhook(
			ctx,
			vcsConnector.Payload.ExternalId,
			webhookID,
		); err != nil {
			slog.Error("failed to delete webhook for VCS connector", slog.String("project", projectID), slog.String("VCS connector", vcsConnector.ResourceID), log.BBError(err))
		}
	}

	return &emptypb.Empty{}, nil
}

func convertStoreVCSConnector(vcsConnector *store.VCSConnectorMessage) (*v1pb.VCSConnector, error) {
	v1VCSConnector := &v1pb.VCSConnector{
		Name:          fmt.Sprintf("%s/%s%s", common.FormatProject(vcsConnector.ProjectID), common.VCSConnectorPrefix, vcsConnector.ResourceID),
		Title:         vcsConnector.Payload.Title,
		VcsProvider:   fmt.Sprintf("%s%s", common.VCSProviderPrefix, vcsConnector.VCSResourceID),
		ExternalId:    vcsConnector.Payload.ExternalId,
		BaseDirectory: vcsConnector.Payload.BaseDirectory,
		Branch:        vcsConnector.Payload.Branch,
		FullPath:      vcsConnector.Payload.FullPath,
		WebUrl:        vcsConnector.Payload.WebUrl,
		DatabaseGroup: vcsConnector.Payload.DatabaseGroup,
	}
	return v1VCSConnector, nil
}

func checkBranchExistence(ctx context.Context, vcsProvider *store.VCSProviderMessage, externalID, branch string) error {
	if branch == "" {
		return status.Errorf(codes.InvalidArgument, "branch name is required")
	}
	if _, err := vcs.Get(vcsProvider.Type, vcs.ProviderConfig{InstanceURL: vcsProvider.InstanceURL, AuthToken: vcsProvider.AccessToken}).GetBranch(ctx, externalID, branch); err != nil {
		if common.ErrorCode(err) == common.NotFound {
			return status.Errorf(codes.NotFound, "branch %s not found in repository %s", branch, externalID)
		}
		return status.Errorf(codes.Internal, "failed to check branch: %v", err)
	}
	return nil
}

func createVCSWebhook(ctx context.Context, vcsProvider *store.VCSProviderMessage, webhookEndpointID, webhookSecretToken, externalRepoID, bytebaseEndpointURL string) (string, error) {
	// Create a new webhook and retrieve the created webhook ID
	var webhookCreatePayloads [][]byte
	switch vcsProvider.Type {
	case storepb.VCSType_GITLAB:
		// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#push-events
		webhookCreate := gitlab.WebhookCreate{
			URL:                   fmt.Sprintf("%s/hook/%s", bytebaseEndpointURL, webhookEndpointID),
			SecretToken:           webhookSecretToken,
			MergeRequestsEvents:   true,
			NoteEvents:            true,
			EnableSSLVerification: false,
		}
		createPayload, err := json.Marshal(webhookCreate)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal request body for creating webhook")
		}
		webhookCreatePayloads = append(webhookCreatePayloads, createPayload)
	case storepb.VCSType_GITHUB:
		webhookPost := github.WebhookCreateOrUpdate{
			Config: github.WebhookConfig{
				URL:         fmt.Sprintf("%s/hook/%s", bytebaseEndpointURL, webhookEndpointID),
				ContentType: "json",
				Secret:      webhookSecretToken,
				InsecureSSL: 1,
			},
			// https://docs.github.com/en/webhooks/webhook-events-and-payloads
			Events: []string{"pull_request", "pull_request_review_comment"},
		}
		createPayload, err := json.Marshal(webhookPost)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal request body for creating webhook")
		}
		webhookCreatePayloads = append(webhookCreatePayloads, createPayload)
	case storepb.VCSType_BITBUCKET:
		webhookPost := bitbucket.WebhookCreateOrUpdate{
			Description: "Bytebase GitOps",
			URL:         fmt.Sprintf("%s/hook/%s", bytebaseEndpointURL, webhookEndpointID),
			Active:      true,
			// https://support.atlassian.com/bitbucket-cloud/docs/event-payloads
			Events: []string{
				string(bitbucket.PullRequestEventCreated),
				string(bitbucket.PullRequestEventUpdated),
				string(bitbucket.PullRequestEventFulfilled),
				"pullrequest:comment_created",
			},
		}
		createPayload, err := json.Marshal(webhookPost)
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal request body for creating webhook")
		}
		webhookCreatePayloads = append(webhookCreatePayloads, createPayload)
	case storepb.VCSType_AZURE_DEVOPS:
		part := strings.Split(externalRepoID, "/")
		if len(part) != 3 {
			return "", errors.Errorf("invalid external repo id %q", externalRepoID)
		}
		projectID, repositoryID := part[1], part[2]

		// https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops
		// Azure doesn't support multiply events in a single webhook, but we need:
		// - git.pullrequest.merged: A merge commit was created on a pull request.
		// - git.pullrequest.created: A pull request is created in a Git repository.
		// - git.pullrequest.updated: A pull request is updated; status, review list, reviewer vote changed, or the source branch is updated with a push.
		events := []azure.PullRequestEventType{
			azure.PullRequestEventCreated,
			azure.PullRequestEventUpdated,
			azure.PullRequestEventMerged,
		}
		for _, event := range events {
			publisherInputs := azure.WebhookCreatePublisherInputs{
				Repository: repositoryID,
				Branch:     "", /* Any branches */
				ProjectID:  projectID,
			}
			if event == azure.PullRequestEventMerged {
				publisherInputs.MergeResult = azure.WebhookMergeResultSucceeded
			}
			webhookPost := azure.WebhookCreateOrUpdate{
				ConsumerActionID: "httpRequest",
				ConsumerID:       "webHooks",
				ConsumerInputs: azure.WebhookCreateConsumerInputs{
					URL:                  fmt.Sprintf("%s/hook/%s", bytebaseEndpointURL, webhookEndpointID),
					AcceptUntrustedCerts: true,
					HTTPHeaders:          fmt.Sprintf("X-Azure-Token: %s", webhookSecretToken),
				},
				EventType:       string(event),
				PublisherID:     "tfs",
				PublisherInputs: publisherInputs,
			}
			createPayload, err := json.Marshal(webhookPost)
			if err != nil {
				return "", errors.Wrap(err, "failed to marshal request body for creating webhook")
			}
			webhookCreatePayloads = append(webhookCreatePayloads, createPayload)
		}
	}

	var webhookIDs []string
	for _, webhookCreatePayload := range webhookCreatePayloads {
		webhookID, err := vcs.Get(vcsProvider.Type, vcs.ProviderConfig{InstanceURL: vcsProvider.InstanceURL, AuthToken: vcsProvider.AccessToken}).CreateWebhook(
			ctx,
			externalRepoID,
			webhookCreatePayload,
		)
		if err != nil {
			return "", errors.Wrap(err, "failed to create webhook")
		}
		webhookIDs = append(webhookIDs, webhookID)
	}
	return strings.Join(webhookIDs, ","), nil
}
