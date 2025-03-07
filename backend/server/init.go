package server

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/bytebase/bytebase/backend/common"
	api "github.com/bytebase/bytebase/backend/legacyapi"
	"github.com/bytebase/bytebase/backend/metric"
	metriccollector "github.com/bytebase/bytebase/backend/metric/collector"
	"github.com/bytebase/bytebase/backend/runner/metricreport"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/backend/store/model"
	storepb "github.com/bytebase/bytebase/proto/generated-go/store"
)

func (s *Server) getInitSetting(ctx context.Context) (string, error) {
	// secretLength is the length for the secret used to sign the JWT auto token.
	const secretLength = 32

	// initial branding
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingBrandingLogo,
		Value: "",
	}); err != nil {
		return "", err
	}

	// initial JWT token
	secret, err := common.RandomString(secretLength)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate random JWT secret")
	}
	authSetting, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingAuthSecret,
		Value: secret,
	})
	if err != nil {
		return "", err
	}
	// Set secret to the stored secret.
	secret = authSetting.Value

	// initial workspace
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingWorkspaceID,
		Value: uuid.New().String(),
	}); err != nil {
		return "", err
	}

	// Init SCIM config
	scimToken, err := common.RandomString(secretLength)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate random SCIM secret")
	}
	scimSettingValue, err := protojson.Marshal(&storepb.SCIMSetting{
		Token: scimToken,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial scim setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingSCIM,
		Value: string(scimSettingValue),
	}); err != nil {
		return "", err
	}

	// Init password validation
	passwordSettingValue, err := protojson.Marshal(&storepb.PasswordRestrictionSetting{
		MinLength:                         8,
		RequireNumber:                     false,
		RequireLetter:                     true,
		RequireUppercaseLetter:            false,
		RequireSpecialCharacter:           false,
		RequireResetPasswordForFirstLogin: false,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial password validation setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingPasswordRestriction,
		Value: string(passwordSettingValue),
	}); err != nil {
		return "", err
	}

	// initial license
	if _, _, err = s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingEnterpriseLicense,
		Value: "",
	}); err != nil {
		return "", err
	}

	// initial IM app
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingAppIM,
		Value: "{}",
	}); err != nil {
		return "", err
	}

	// initial watermark setting
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingWatermark,
		Value: "0",
	}); err != nil {
		return "", err
	}

	// initial OpenAI key setting
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingPluginOpenAIKey,
		Value: "",
	}); err != nil {
		return "", err
	}

	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingPluginOpenAIEndpoint,
		Value: "",
	}); err != nil {
		return "", err
	}

	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingPluginOpenAIModel,
		Value: "",
	}); err != nil {
		return "", err
	}

	// initial external approval setting
	externalApprovalSettingValue, err := protojson.Marshal(&storepb.ExternalApprovalSetting{})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial external approval setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingWorkspaceExternalApproval,
		Value: string(externalApprovalSettingValue),
	}); err != nil {
		return "", err
	}

	// initial schema template setting
	schemaTemplateSettingValue, err := protojson.Marshal(&storepb.SchemaTemplateSetting{})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial schema template setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingSchemaTemplate,
		Value: string(schemaTemplateSettingValue),
	}); err != nil {
		return "", err
	}

	// initial data classification setting
	dataClassificationSettingValue, err := protojson.Marshal(&storepb.DataClassificationSetting{})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial data classification setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name:  api.SettingDataClassification,
		Value: string(dataClassificationSettingValue),
	}); err != nil {
		return "", err
	}

	// initial workspace approval setting
	approvalSettingValue, err := protojson.Marshal(&storepb.WorkspaceApprovalSetting{})
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal initial workspace approval setting")
	}
	if _, _, err := s.store.CreateSettingIfNotExistV2(ctx, &store.SettingMessage{
		Name: api.SettingWorkspaceApproval,
		// Value is ""
		Value: string(approvalSettingValue),
	}); err != nil {
		return "", err
	}

	// initial workspace profile setting
	workspaceProfileSetting, err := s.store.GetSettingV2(ctx, api.SettingWorkspaceProfile)
	if err != nil {
		return "", err
	}

	workspaceProfilePayload := &storepb.WorkspaceProfileSetting{
		ExternalUrl: s.profile.ExternalURL,
	}
	if workspaceProfileSetting != nil {
		workspaceProfilePayload = new(storepb.WorkspaceProfileSetting)
		if err := common.ProtojsonUnmarshaler.Unmarshal([]byte(workspaceProfileSetting.Value), workspaceProfilePayload); err != nil {
			return "", err
		}
		if s.profile.ExternalURL != "" {
			workspaceProfilePayload.ExternalUrl = s.profile.ExternalURL
		}
	}

	bytes, err := protojson.Marshal(workspaceProfilePayload)
	if err != nil {
		return "", err
	}

	if _, err := s.store.UpsertSettingV2(ctx, &store.SetSettingMessage{
		Name:  api.SettingWorkspaceProfile,
		Value: string(bytes),
	}); err != nil {
		return "", err
	}

	// Init workspace IAM policy
	if _, err := s.store.PatchWorkspaceIamPolicy(ctx, &store.PatchIamPolicyMessage{
		Member: common.FormatUserUID(api.SystemBotID),
		Roles: []string{
			common.FormatRole(api.WorkspaceAdmin.String()),
		},
	}); err != nil {
		return "", err
	}
	if _, err := s.store.PatchWorkspaceIamPolicy(ctx, &store.PatchIamPolicyMessage{
		Member: api.AllUsers,
		Roles: []string{
			common.FormatRole(api.WorkspaceMember.String()),
		},
	}); err != nil {
		return "", err
	}

	return secret, nil
}

func (s *Server) migrateMaskingData(ctx context.Context) error {
	resourceType := api.PolicyResourceTypeDatabase
	policyType := api.PolicyTypeMasking
	policies, err := s.store.ListPoliciesV2(ctx, &store.FindPolicyMessage{
		ResourceType: &resourceType,
		Type:         &policyType,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list masking policy")
	}

	if len(policies) > 0 {
		slog.Info("Begin migrate database masking policy...")
	}

	for _, policy := range policies {
		p := new(storepb.DeprecatedMaskingPolicy)
		if err := common.ProtojsonUnmarshaler.Unmarshal([]byte(policy.Payload), p); err != nil {
			return errors.Wrapf(err, "failed to unmarshal masking policy")
		}

		dbSchema, err := s.store.GetDBSchema(ctx, policy.ResourceUID)
		if err != nil {
			return errors.Wrapf(err, "failed to get schema for database %v", policy.ResourceUID)
		}
		dbModelConfig := model.NewDatabaseConfig(nil)
		if dbSchema != nil {
			dbModelConfig = dbSchema.GetInternalConfig()
		}
		for _, mask := range p.MaskData {
			schemaConfig := dbModelConfig.CreateOrGetSchemaConfig(mask.Schema)
			tableConfig := schemaConfig.CreateOrGetTableConfig(mask.Table)
			columnConfig := tableConfig.CreateOrGetColumnConfig(mask.Column)
			if mask.FullMaskingAlgorithmId != "" {
				columnConfig.SemanticType = mask.FullMaskingAlgorithmId
			} else if mask.PartialMaskingAlgorithmId != "" {
				columnConfig.SemanticType = mask.PartialMaskingAlgorithmId
			}
		}

		if err := s.store.UpdateDBSchema(ctx, policy.ResourceUID, &store.UpdateDBSchemaMessage{Config: dbModelConfig.BuildDatabaseConfig()}); err != nil {
			return errors.Wrapf(err, "failed to update db config for database %v", policy.ResourceUID)
		}
		if err := s.store.DeletePolicyV2(ctx, &store.PolicyMessage{
			ResourceUID:  policy.ResourceUID,
			ResourceType: resourceType,
			Type:         policyType,
		}); err != nil {
			return errors.Wrapf(err, "failed to delete legacy masking policy for database %v", policy.ResourceUID)
		}
	}
	if len(policies) > 0 {
		slog.Info("Database masking policy migration finished.")
	}
	return nil
}

func (s *Server) migrateCatalog(ctx context.Context) error {
	databaseIDs, err := s.store.ListLegacyCatalog(ctx)
	if err != nil {
		return err
	}
	for _, databaseID := range databaseIDs {
		dbSchema, err := s.store.GetDBSchema(ctx, databaseID)
		if err != nil {
			return errors.Wrapf(err, "failed to get schema for database %v", databaseID)
		}
		if dbSchema == nil {
			continue
		}
		if dbSchema.GetConfig() == nil {
			continue
		}
		cfg := dbSchema.GetConfig()
		updated := false
		for _, s := range cfg.Schemas {
			for _, t := range s.Tables {
				for _, col := range t.Columns {
					switch col.GetMaskingLevel() {
					case storepb.MaskingLevel_FULL:
						if col.GetFullMaskingAlgorithmId() != "" {
							col.SemanticType = col.GetFullMaskingAlgorithmId()
						} else {
							col.SemanticType = "bb.default"
						}
						updated = true
					case storepb.MaskingLevel_PARTIAL:
						if col.GetPartialMaskingAlgorithmId() != "" {
							col.SemanticType = col.GetPartialMaskingAlgorithmId()
						} else {
							col.SemanticType = "bb.default-partial"
						}
						updated = true
					}
					col.MaskingLevel = storepb.MaskingLevel_MASKING_LEVEL_UNSPECIFIED
					col.FullMaskingAlgorithmId = ""
					col.PartialMaskingAlgorithmId = ""
				}
			}
		}
		if updated {
			if err := s.store.UpdateDBSchema(ctx, databaseID, &store.UpdateDBSchemaMessage{Config: cfg}); err != nil {
				return errors.Wrapf(err, "failed to update db config for database %v", databaseID)
			}
			slog.Info("migrated catalog for database", slog.Int("databaseID", databaseID))
		}
	}
	return nil
}

// initMetricReporter will initial the metric scheduler.
func (s *Server) initMetricReporter() {
	metricReporter := metricreport.NewReporter(s.store, s.licenseService, s.profile, s.profile.EnableMetric)
	metricReporter.Register(metric.InstanceCountMetricName, metriccollector.NewInstanceCountCollector(s.store))
	metricReporter.Register(metric.IssueCountMetricName, metriccollector.NewIssueCountCollector(s.store))
	metricReporter.Register(metric.ProjectCountMetricName, metriccollector.NewProjectCountCollector(s.store))
	metricReporter.Register(metric.PolicyCountMetricName, metriccollector.NewPolicyCountCollector(s.store))
	metricReporter.Register(metric.TaskCountMetricName, metriccollector.NewTaskCountCollector(s.store))
	metricReporter.Register(metric.SheetCountMetricName, metriccollector.NewSheetCountCollector(s.store))
	metricReporter.Register(metric.MemberCountMetricName, metriccollector.NewMemberCountCollector(s.store))
	s.metricReporter = metricReporter
}
