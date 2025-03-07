package taskrun

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/common/log"
	"github.com/bytebase/bytebase/backend/component/config"
	"github.com/bytebase/bytebase/backend/component/dbfactory"
	"github.com/bytebase/bytebase/backend/component/state"
	api "github.com/bytebase/bytebase/backend/legacyapi"
	"github.com/bytebase/bytebase/backend/plugin/db"
	"github.com/bytebase/bytebase/backend/runner/schemasync"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/backend/store/model"
	"github.com/bytebase/bytebase/backend/utils"
	storepb "github.com/bytebase/bytebase/proto/generated-go/store"
)

// Executor is the task executor.
type Executor interface {
	// RunOnce will be called periodically by the scheduler until terminated is true.
	//
	// NOTE
	//
	// 1. It's possible that err could be non-nil while terminated is false, which
	// usually indicates a transient error and will make scheduler retry later.
	// 2. If err is non-nil, then the detail field will be ignored since info is provided in the err.
	// driverCtx is used by the database driver so that we can cancel the query
	// while have the ability to cleanup migration history etc.
	RunOnce(ctx context.Context, driverCtx context.Context, task *store.TaskMessage, taskRunUID int) (terminated bool, result *storepb.TaskRunResult, err error)
}

// RunExecutorOnce wraps a TaskExecutor.RunOnce call with panic recovery.
func RunExecutorOnce(ctx context.Context, driverCtx context.Context, exec Executor, task *store.TaskMessage, taskRunUID int) (terminated bool, result *storepb.TaskRunResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr, ok := r.(error)
			if !ok {
				panicErr = errors.Errorf("%v", r)
			}
			slog.Error("TaskExecutor PANIC RECOVER", log.BBError(panicErr), log.BBStack("panic-stack"))
			terminated = true
			result = nil
			err = errors.Errorf("TaskExecutor PANIC RECOVER, err: %v", panicErr)
		}
	}()

	return exec.RunOnce(ctx, driverCtx, task, taskRunUID)
}

// Pointer fields are not nullable unless mentioned otherwise.
type migrateContext struct {
	syncer    *schemasync.Syncer
	profile   *config.Profile
	dbFactory *dbfactory.DBFactory

	instance *store.InstanceMessage
	database *store.DatabaseMessage
	// nullable if type=baseline
	sheet *store.SheetMessage
	// empty if type=baseline
	sheetName string

	task        *store.TaskMessage
	taskRunUID  int
	taskRunName string
	issueName   string

	version string

	release struct {
		// The release
		// Format: projects/{project}/releases/{release}
		release string
		// The file
		// Format: projects/{project}/releases/{release}/files/{id}
		file string
	}

	// mutable
	// changelog uid
	changelog int64
}

func getMigrationInfo(ctx context.Context, stores *store.Store, profile *config.Profile, syncer *schemasync.Syncer, task *store.TaskMessage, migrationType db.MigrationType, statement string, schemaVersion model.Version, sheetID *int, taskRunUID int, dbFactory *dbfactory.DBFactory) (*db.MigrationInfo, *migrateContext, error) {
	instance, err := stores.GetInstanceV2(ctx, &store.FindInstanceMessage{UID: &task.InstanceID})
	if err != nil {
		return nil, nil, err
	}
	database, err := stores.GetDatabaseV2(ctx, &store.FindDatabaseMessage{UID: task.DatabaseID})
	if err != nil {
		return nil, nil, err
	}
	if database == nil {
		return nil, nil, errors.Errorf("database not found")
	}
	environment, err := stores.GetEnvironmentV2(ctx, &store.FindEnvironmentMessage{ResourceID: &database.EffectiveEnvironmentID})
	if err != nil {
		return nil, nil, err
	}

	mi := &db.MigrationInfo{
		InstanceID:     &instance.UID,
		DatabaseID:     &database.UID,
		ReleaseVersion: profile.Version,
		Type:           migrationType,
		Description:    task.Name,
		Environment:    environment.ResourceID,
		Database:       database.DatabaseName,
		Namespace:      database.DatabaseName,
		Payload:        &storepb.InstanceChangeHistoryPayload{},
	}

	pipeline, err := stores.GetPipelineV2ByID(ctx, task.PipelineID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get pipeline")
	}
	if pipeline == nil {
		return nil, nil, errors.Errorf("pipeline %v not found", task.PipelineID)
	}

	mc := &migrateContext{
		syncer:      syncer,
		profile:     profile,
		dbFactory:   dbFactory,
		instance:    instance,
		database:    database,
		task:        task,
		version:     schemaVersion.Version,
		taskRunName: common.FormatTaskRun(pipeline.ProjectID, task.PipelineID, task.StageID, task.ID, taskRunUID),
		taskRunUID:  taskRunUID,
	}

	if sheetID != nil {
		sheet, err := stores.GetSheet(ctx, &store.FindSheetMessage{UID: sheetID})
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get sheet")
		}
		if sheet == nil {
			return nil, nil, errors.Errorf("sheet not found")
		}
		mc.sheet = sheet
		mc.sheetName = common.FormatSheet(pipeline.ProjectID, sheet.UID)
	}

	if task.Type.ChangeDatabasePayload() {
		var p storepb.TaskDatabaseUpdatePayload
		if err := common.ProtojsonUnmarshaler.Unmarshal([]byte(task.Payload), &p); err != nil {
			return nil, nil, errors.Wrapf(err, "failed to unmarshal task payload")
		}

		if f := p.TaskReleaseSource.GetFile(); f != "" {
			project, release, _, err := common.GetProjectReleaseUIDFile(f)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to parse file %s", f)
			}
			mc.release.release = common.FormatReleaseName(project, release)
			mc.release.file = f
		}
	}

	plans, err := stores.ListPlans(ctx, &store.FindPlanMessage{PipelineID: &task.PipelineID})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to list plans")
	}
	if len(plans) == 1 {
		planTypes := []store.PlanCheckRunType{store.PlanCheckDatabaseStatementSummaryReport}
		status := []store.PlanCheckRunStatus{store.PlanCheckRunStatusDone}
		runs, err := stores.ListPlanCheckRuns(ctx, &store.FindPlanCheckRunMessage{
			PlanUID: &plans[0].UID,
			Type:    &planTypes,
			Status:  &status,
		})
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to list plan check runs")
		}
		sort.Slice(runs, func(i, j int) bool {
			return runs[i].UID > runs[j].UID
		})
		foundChangedResources := false
		for _, run := range runs {
			if foundChangedResources {
				break
			}
			if run.Config.InstanceUid != int32(task.InstanceID) {
				continue
			}
			if run.Config.DatabaseName != database.DatabaseName {
				continue
			}
			if sheetID != nil && run.Config.SheetUid != int32(*sheetID) {
				continue
			}
			if run.Result == nil {
				continue
			}
			for _, result := range run.Result.Results {
				if result.Status != storepb.PlanCheckRunResult_Result_SUCCESS {
					continue
				}
				if report := result.GetSqlSummaryReport(); report != nil {
					mi.Payload.ChangedResources = report.ChangedResources
					foundChangedResources = true
					break
				}
			}
		}
	}

	issue, err := stores.GetIssueV2(ctx, &store.FindIssueMessage{PipelineID: &task.PipelineID})
	if err != nil {
		slog.Error("failed to find containing issue", log.BBError(err))
	}
	if issue != nil {
		// Concat issue title and task name as the migration description so that user can see
		// more context of the migration.
		mi.Description = fmt.Sprintf("%s - %s", issue.Title, task.Name)
		mi.ProjectUID = &issue.Project.UID
		mi.IssueUID = &issue.UID

		mc.issueName = common.FormatIssue(issue.Project.ResourceID, issue.UID)
	}

	mi.Source = db.UI

	statement = strings.TrimSpace(statement)
	// Only baseline and SDL migration can have empty sql statement, which indicates empty database.
	if mi.Type != db.Baseline && mi.Type != db.MigrateSDL && statement == "" {
		return nil, nil, errors.Errorf("empty statement")
	}
	return mi, mc, nil
}

func getCreateTaskRunLog(ctx context.Context, taskRunUID int, s *store.Store, profile *config.Profile) func(t time.Time, e *storepb.TaskRunLog) error {
	return func(t time.Time, e *storepb.TaskRunLog) error {
		return s.CreateTaskRunLog(ctx, taskRunUID, t.UTC(), profile.DeployID, e)
	}
}

func getUseDatabaseOwner(ctx context.Context, stores *store.Store, instance *store.InstanceMessage, database *store.DatabaseMessage) (bool, error) {
	if instance.Engine != storepb.Engine_POSTGRES {
		return false, nil
	}

	// Check the project setting to see if we should use the database owner.
	project, err := stores.GetProjectV2(ctx, &store.FindProjectMessage{ResourceID: &database.ProjectID})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get project")
	}

	if project.Setting == nil {
		return false, nil
	}

	return project.Setting.PostgresDatabaseTenantMode, nil
}

func doMigration(
	ctx context.Context,
	driverCtx context.Context,
	stores *store.Store,
	stateCfg *state.State,
	profile *config.Profile,
	statement string,
	mi *db.MigrationInfo,
	mc *migrateContext,
) (bool, error) {
	instance := mc.instance
	database := mc.database

	useDBOwner, err := getUseDatabaseOwner(ctx, stores, instance, database)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if we should use database owner")
	}
	driver, err := mc.dbFactory.GetAdminDatabaseDriver(ctx, instance, database, db.ConnectionContext{
		UseDatabaseOwner: useDBOwner,
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get driver connection for instance %q", instance.ResourceID)
	}
	defer driver.Close(ctx)

	statementRecord, _ := common.TruncateString(statement, common.MaxSheetSize)
	slog.Debug("Start migration...",
		slog.String("instance", instance.ResourceID),
		slog.String("database", database.DatabaseName),
		slog.String("source", string(mi.Source)),
		slog.String("type", string(mi.Type)),
		slog.String("statement", statementRecord),
	)

	opts := db.ExecuteOptions{}

	opts.SetConnectionID = func(id string) {
		stateCfg.TaskRunConnectionID.Store(mc.taskRunUID, id)
	}
	opts.DeleteConnectionID = func() {
		stateCfg.TaskRunConnectionID.Delete(mc.taskRunUID)
	}

	if stateCfg != nil {
		switch mc.task.Type {
		case api.TaskDatabaseSchemaUpdate, api.TaskDatabaseDataUpdate:
			switch instance.Engine {
			case storepb.Engine_MYSQL, storepb.Engine_TIDB, storepb.Engine_OCEANBASE,
				storepb.Engine_STARROCKS, storepb.Engine_DORIS, storepb.Engine_POSTGRES,
				storepb.Engine_REDSHIFT, storepb.Engine_RISINGWAVE, storepb.Engine_ORACLE,
				storepb.Engine_DM, storepb.Engine_OCEANBASE_ORACLE, storepb.Engine_MSSQL,
				storepb.Engine_DYNAMODB:
				opts.CreateTaskRunLog = getCreateTaskRunLog(ctx, mc.taskRunUID, stores, profile)
			default:
				// do nothing
			}
		}
	}

	return executeMigrationDefault(ctx, driverCtx, stores, stateCfg, driver, mi, mc, statement, opts)
}

func postMigration(ctx context.Context, stores *store.Store, mi *db.MigrationInfo, mc *migrateContext, skipped bool) (bool, *storepb.TaskRunResult, error) {
	if skipped {
		return true, &storepb.TaskRunResult{
			Detail: fmt.Sprintf("Task skipped because version %s has been applied", mc.version),
		}, nil
	}

	instance := mc.instance
	database := mc.database

	slog.Debug("Post migration...",
		slog.String("instance", instance.ResourceID),
		slog.String("database", database.DatabaseName),
	)

	// Remove schema drift anomalies.
	if err := stores.DeleteAnomalyV2(ctx, &store.DeleteAnomalyMessage{
		DatabaseUID: *(mc.task.DatabaseID),
		Type:        api.AnomalyDatabaseSchemaDrift,
	}); err != nil && common.ErrorCode(err) != common.NotFound {
		slog.Error("Failed to archive anomaly",
			slog.String("instance", instance.ResourceID),
			slog.String("database", database.DatabaseName),
			slog.String("type", string(api.AnomalyDatabaseSchemaDrift)),
			log.BBError(err))
	}

	detail := fmt.Sprintf("Applied migration version %s to database %q.", mc.version, database.DatabaseName)
	if mi.Type == db.Baseline {
		detail = fmt.Sprintf("Established baseline version %s for database %q.", mc.version, database.DatabaseName)
	}

	return true, &storepb.TaskRunResult{
		Detail:    detail,
		Changelog: common.FormatChangelog(instance.ResourceID, database.DatabaseName, mc.changelog),
		Version:   mc.version,
	}, nil
}

func runMigration(ctx context.Context, driverCtx context.Context, store *store.Store, dbFactory *dbfactory.DBFactory, stateCfg *state.State, syncer *schemasync.Syncer, profile *config.Profile, task *store.TaskMessage, taskRunUID int, migrationType db.MigrationType, statement string, schemaVersion model.Version, sheetID *int) (terminated bool, result *storepb.TaskRunResult, err error) {
	mi, mc, err := getMigrationInfo(ctx, store, profile, syncer, task, migrationType, statement, schemaVersion, sheetID, taskRunUID, dbFactory)
	if err != nil {
		return true, nil, err
	}

	skipped, err := doMigration(ctx, driverCtx, store, stateCfg, profile, statement, mi, mc)
	if err != nil {
		return true, nil, err
	}
	return postMigration(ctx, store, mi, mc, skipped)
}

// executeMigrationDefault executes migration.
func executeMigrationDefault(ctx context.Context, driverCtx context.Context, store *store.Store, _ *state.State, driver db.Driver, mi *db.MigrationInfo, mc *migrateContext, statement string, opts db.ExecuteOptions) (skipped bool, resErr error) {
	execFunc := func(ctx context.Context, execStatement string) error {
		if _, err := driver.Execute(ctx, execStatement, opts); err != nil {
			return err
		}
		return nil
	}
	return executeMigrationWithFunc(ctx, driverCtx, store, mi, mc, statement, execFunc, opts)
}

// executeMigrationWithFunc executes the migration with custom migration function.
func executeMigrationWithFunc(ctx context.Context, driverCtx context.Context, s *store.Store, mi *db.MigrationInfo, mc *migrateContext, statement string, execFunc func(ctx context.Context, execStatement string) error, opts db.ExecuteOptions) (skipped bool, resErr error) {
	// Phase 1 - Dump before migration.
	// Check if versioned is already applied.
	skipExecution, err := beginMigration(ctx, s, mi, mc, opts)
	if err != nil {
		return false, errors.Wrapf(err, "failed to begin migration")
	}
	if skipExecution {
		return true, nil
	}

	defer func() {
		// Phase 3 - Dump after migration.
		// Insert revision for versioned.
		if err := endMigration(ctx, s, mi, mc, resErr == nil /* isDone */); err != nil {
			slog.Error("failed to end migration",
				log.BBError(err),
			)
		}
	}()

	// Phase 2 - Executing migration.
	// Branch migration type always has empty sql.
	// Baseline migration type could has non-empty sql but will not execute.
	// https://github.com/bytebase/bytebase/issues/394
	doMigrate := true
	if statement == "" || mi.Type == db.Baseline {
		doMigrate = false
	}
	if doMigrate {
		renderedStatement := statement
		// The m.DatabaseID is nil means the migration is a instance level migration
		if mi.DatabaseID != nil {
			database, err := s.GetDatabaseV2(ctx, &store.FindDatabaseMessage{
				UID: mi.DatabaseID,
			})
			if err != nil {
				return false, err
			}
			if database == nil {
				return false, errors.Errorf("database %d not found", *mi.DatabaseID)
			}
			materials := utils.GetSecretMapFromDatabaseMessage(database)
			// To avoid leak the rendered statement, the error message should use the original statement and not the rendered statement.
			renderedStatement = utils.RenderStatement(statement, materials)
		}

		if err := execFunc(driverCtx, renderedStatement); err != nil {
			return false, err
		}
	}

	return false, nil
}

// beginMigration checks before executing migration and inserts a migration history record with pending status.
func beginMigration(ctx context.Context, stores *store.Store, mi *db.MigrationInfo, mc *migrateContext, opts db.ExecuteOptions) (bool, error) {
	// list revisions and see if it has been applied
	// we can do this because
	// versioned migrations are executed one by one
	// so no other migrations can insert revisions
	//
	// users can create revisions though via API
	// however we can warn users not to unless they know
	// what they are doing
	if mc.version != "" {
		list, err := stores.ListRevisions(ctx, &store.FindRevisionMessage{
			DatabaseUID: &mc.database.UID,
			Version:     &mc.version,
		})
		if err != nil {
			return false, errors.Wrapf(err, "failed to list revisions")
		}
		if len(list) > 0 {
			// This version has been executed.
			// skip execution.
			return true, nil
		}
	}

	// sync history
	var syncHistoryPrevUID *int64
	if mi.Type.NeedDump() {
		opts.LogDatabaseSyncStart()
		syncHistoryPrev, err := mc.syncer.SyncDatabaseSchemaToHistory(ctx, mc.database, false)
		if err != nil {
			opts.LogDatabaseSyncEnd(err.Error())
			return false, errors.Wrapf(err, "failed to sync database metadata and schema")
		}
		opts.LogDatabaseSyncEnd("")
		syncHistoryPrevUID = &syncHistoryPrev
	}

	// create pending changelog
	changelogUID, err := stores.CreateChangelog(ctx, &store.ChangelogMessage{
		DatabaseUID:        mc.database.UID,
		Status:             store.ChangelogStatusPending,
		PrevSyncHistoryUID: syncHistoryPrevUID,
		SyncHistoryUID:     nil,
		Payload: &storepb.ChangelogPayload{
			TaskRun:          mc.taskRunName,
			Issue:            mc.issueName,
			Revision:         0,
			ChangedResources: mi.Payload.GetChangedResources(),
			Sheet:            mc.sheetName,
			Version:          mc.version,
			Type:             convertTaskType(mc.task.Type),
		}})
	if err != nil {
		return false, errors.Wrapf(err, "failed to create changelog")
	}
	mc.changelog = changelogUID

	return false, nil
}

// endMigration updates the migration history record to DONE or FAILED depending on migration is done or not.
func endMigration(ctx context.Context, storeInstance *store.Store, mi *db.MigrationInfo, mc *migrateContext, isDone bool) error {
	update := &store.UpdateChangelogMessage{
		UID: mc.changelog,
	}

	if mi.Type.NeedDump() {
		syncHistory, err := mc.syncer.SyncDatabaseSchemaToHistory(ctx, mc.database, false)
		if err != nil {
			return errors.Wrapf(err, "failed to sync database metadata and schema")
		}
		update.SyncHistoryUID = &syncHistory
	}

	if isDone {
		// if isDone, record in revision
		if mc.version != "" {
			r := &store.RevisionMessage{
				DatabaseUID: mc.database.UID,
				Version:     mc.version,
				Payload: &storepb.RevisionPayload{
					Release:     mc.release.release,
					File:        mc.release.file,
					Sheet:       "",
					SheetSha256: "",
					TaskRun:     mc.taskRunName,
				},
			}
			if mc.sheet != nil {
				r.Payload.Sheet = mc.sheetName
				r.Payload.SheetSha256 = mc.sheet.GetSha256Hex()
			}

			revision, err := storeInstance.CreateRevision(ctx, r)
			if err != nil {
				return errors.Wrapf(err, "failed to create revision")
			}
			update.RevisionUID = &revision.UID
		}
		status := store.ChangelogStatusDone
		update.Status = &status
	} else {
		status := store.ChangelogStatusFailed
		update.Status = &status
	}

	if err := storeInstance.UpdateChangelog(ctx, update); err != nil {
		return errors.Wrapf(err, "failed to update changelog")
	}

	return nil
}

func convertTaskType(t api.TaskType) storepb.ChangelogPayload_Type {
	switch t {
	case api.TaskDatabaseDataUpdate:
		return storepb.ChangelogPayload_DATA
	case api.TaskDatabaseSchemaBaseline:
		return storepb.ChangelogPayload_BASELINE
	case api.TaskDatabaseSchemaUpdate:
		return storepb.ChangelogPayload_MIGRATE
	case api.TaskDatabaseSchemaUpdateSDL:
		return storepb.ChangelogPayload_MIGRATE_SDL
	case api.TaskDatabaseSchemaUpdateGhostCutover, api.TaskDatabaseSchemaUpdateGhostSync:
		return storepb.ChangelogPayload_MIGRATE_GHOST

	case api.TaskGeneral:
	case api.TaskDatabaseCreate:
	case api.TaskDatabaseDataExport:
	}
	return storepb.ChangelogPayload_TYPE_UNSPECIFIED
}
