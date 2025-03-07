-- Type
CREATE TYPE row_status AS ENUM ('NORMAL', 'ARCHIVED');

-- idp stores generic identity provider.
CREATE TABLE idp (
  id SERIAL PRIMARY KEY,
  row_status row_status NOT NULL DEFAULT 'NORMAL',
  resource_id TEXT NOT NULL,
  name TEXT NOT NULL,
  domain TEXT NOT NULL,
  type TEXT NOT NULL CONSTRAINT idp_type_check CHECK (type IN ('OAUTH2', 'OIDC', 'LDAP')),
  -- config stores the corresponding configuration of the IdP, which may vary depending on the type of the IdP.
  config JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_idp_unique_resource_id ON idp(resource_id);

ALTER SEQUENCE idp_id_seq RESTART WITH 101;

-- principal
CREATE TABLE principal (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    type TEXT NOT NULL CHECK (type IN ('END_USER', 'SYSTEM_BOT', 'SERVICE_ACCOUNT')),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    mfa_config JSONB NOT NULL DEFAULT '{}',
    profile JSONB NOT NULL DEFAULT '{}'
);

-- Setting
CREATE TABLE setting (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_setting_unique_name ON setting(name);

ALTER SEQUENCE setting_id_seq RESTART WITH 101;

-- Role
CREATE TABLE role (
    id BIGSERIAL PRIMARY KEY,
    resource_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '{}',
    payload JSONB NOT NULL DEFAULT '{}' -- saved for future use
);

CREATE UNIQUE INDEX idx_role_unique_resource_id on role (resource_id);

ALTER SEQUENCE role_id_seq RESTART WITH 101;

-- Environment
CREATE TABLE environment (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    name TEXT NOT NULL,
    "order" INTEGER NOT NULL CHECK ("order" >= 0),
    resource_id TEXT NOT NULL
);

CREATE UNIQUE INDEX idx_environment_unique_resource_id ON environment(resource_id);

ALTER SEQUENCE environment_id_seq RESTART WITH 101;

-- Policy
-- policy stores the policies for each environment.
-- Policies are associated with environments. Since we may have policies not associated with environment later, we name the table policy.
CREATE TYPE resource_type AS ENUM ('WORKSPACE', 'ENVIRONMENT', 'PROJECT', 'INSTANCE', 'DATABASE');

CREATE TABLE policy (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    type TEXT NOT NULL CHECK (type LIKE 'bb.policy.%'),
    payload JSONB NOT NULL DEFAULT '{}',
    resource_type resource_type NOT NULL,
    resource_id INTEGER NOT NULL,
    inherit_from_parent BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX idx_policy_unique_resource_type_resource_id_type ON policy(resource_type, resource_id, type);

ALTER SEQUENCE policy_id_seq RESTART WITH 101;

-- Project
CREATE TABLE project (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    name TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    data_classification_config_id TEXT NOT NULL DEFAULT '',
    setting JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_project_unique_resource_id ON project(resource_id);

-- Project Hook
CREATE TABLE project_webhook (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project (id),
    type TEXT NOT NULL CHECK (type LIKE 'bb.plugin.webhook.%'),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    activity_list TEXT ARRAY NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_project_webhook_project_id ON project_webhook(project_id);

CREATE UNIQUE INDEX idx_project_webhook_unique_project_id_url ON project_webhook(project_id, url);

ALTER SEQUENCE project_webhook_id_seq RESTART WITH 101;

-- Instance
CREATE TABLE instance (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    environment TEXT REFERENCES environment (resource_id),
    name TEXT NOT NULL,
    engine TEXT NOT NULL,
    engine_version TEXT NOT NULL DEFAULT '',
    external_link TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL,
    -- activation should set to be TRUE if users assign license to this instance.
    activation BOOLEAN NOT NULL DEFAULT false,
    options JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_instance_unique_resource_id ON instance(resource_id);

ALTER SEQUENCE instance_id_seq RESTART WITH 101;

-- db stores the databases for a particular instance
-- data is synced periodically from the instance
CREATE TABLE db (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES instance (id),
    project_id INTEGER NOT NULL REFERENCES project (id),
    environment TEXT REFERENCES environment (resource_id),
    sync_status TEXT NOT NULL CHECK (sync_status IN ('OK', 'NOT_FOUND')),
    sync_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    schema_version TEXT NOT NULL,
    name TEXT NOT NULL,
    secrets JSONB NOT NULL DEFAULT '{}',
    datashare BOOLEAN NOT NULL DEFAULT FALSE,
    -- service_name is the Oracle specific field.
    service_name TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_db_instance_id ON db(instance_id);

CREATE UNIQUE INDEX idx_db_unique_instance_id_name ON db(instance_id, name);

CREATE INDEX idx_db_project_id ON db(project_id);

ALTER SEQUENCE db_id_seq RESTART WITH 101;

-- db_schema stores the database schema metadata for a particular database.
CREATE TABLE db_schema (
    id SERIAL PRIMARY KEY,
    database_id INTEGER NOT NULL REFERENCES db (id) ON DELETE CASCADE,
    metadata JSON NOT NULL DEFAULT '{}',
    raw_dump TEXT NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_db_schema_unique_database_id ON db_schema(database_id);

ALTER SEQUENCE db_schema_id_seq RESTART WITH 101;

-- data_source table stores the data source for a particular database
CREATE TABLE data_source (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES instance (id),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('ADMIN', 'RW', 'RO')),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    ssl_key TEXT NOT NULL DEFAULT '',
    ssl_cert TEXT NOT NULL DEFAULT '',
    ssl_ca TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    port TEXT NOT NULL DEFAULT '',
    options JSONB NOT NULL DEFAULT '{}',
    database TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_data_source_unique_instance_id_name ON data_source(instance_id, name);

ALTER SEQUENCE data_source_id_seq RESTART WITH 101;

CREATE TABLE sheet_blob (
	sha256 BYTEA NOT NULL PRIMARY KEY,
	content TEXT NOT NULL
);

-- sheet table stores general statements.
CREATE TABLE sheet (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    name TEXT NOT NULL,
    sha256 BYTEA NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_sheet_project_id ON sheet(project_id);


ALTER SEQUENCE sheet_id_seq RESTART WITH 101;

-----------------------
-- Pipeline related BEGIN
-- pipeline table
CREATE TABLE pipeline (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    name TEXT NOT NULL
);

ALTER SEQUENCE pipeline_id_seq RESTART WITH 101;

-- stage table stores the stage for the pipeline
CREATE TABLE stage (
    id SERIAL PRIMARY KEY,
    pipeline_id INTEGER NOT NULL REFERENCES pipeline (id),
    environment_id INTEGER NOT NULL REFERENCES environment (id),
    deployment_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL
);

CREATE INDEX idx_stage_pipeline_id ON stage(pipeline_id);

ALTER SEQUENCE stage_id_seq RESTART WITH 101;

-- task table stores the task for the stage
CREATE TABLE task (
    id SERIAL PRIMARY KEY,
    pipeline_id INTEGER NOT NULL REFERENCES pipeline (id),
    stage_id INTEGER NOT NULL REFERENCES stage (id),
    instance_id INTEGER NOT NULL REFERENCES instance (id),
    -- Could be empty for creating database task when the task isn't yet completed successfully.
    database_id INTEGER REFERENCES db (id),
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'PENDING_APPROVAL', 'RUNNING', 'DONE', 'FAILED', 'CANCELED')),
    type TEXT NOT NULL CHECK (type LIKE 'bb.task.%'),
    payload JSONB NOT NULL DEFAULT '{}',
    earliest_allowed_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_task_pipeline_id_stage_id ON task(pipeline_id, stage_id);

CREATE INDEX idx_task_status ON task(status);

ALTER SEQUENCE task_id_seq RESTART WITH 101;

-- task_dag describes task dependency relationship
-- from_task_id blocks to_task_id
CREATE TABLE task_dag (
    id SERIAL PRIMARY KEY,
    from_task_id INTEGER NOT NULL REFERENCES task (id),
    to_task_id INTEGER NOT NULL REFERENCES task (id)
);

CREATE INDEX idx_task_dag_from_task_id ON task_dag(from_task_id);

CREATE INDEX idx_task_dag_to_task_id ON task_dag(to_task_id);

ALTER SEQUENCE task_dag_id_seq RESTART WITH 101;

-- task run table stores the task run
CREATE TABLE task_run (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    task_id INTEGER NOT NULL REFERENCES task (id),
    sheet_id INTEGER REFERENCES sheet (id),
    attempt INTEGER NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'RUNNING', 'DONE', 'FAILED', 'CANCELED')),
    started_at TIMESTAMPTZ NULL,
    code INTEGER NOT NULL DEFAULT 0,
    -- result saves the task run result in json format
    result  JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_task_run_task_id ON task_run(task_id);

CREATE UNIQUE INDEX uk_task_run_task_id_attempt ON task_run (task_id, attempt);

ALTER SEQUENCE task_run_id_seq RESTART WITH 101;

CREATE TABLE task_run_log (
    id BIGSERIAL PRIMARY KEY,
    task_run_id INTEGER NOT NULL REFERENCES task_run (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_task_run_log_task_run_id ON task_run_log(task_run_id);

ALTER SEQUENCE task_run_log_id_seq RESTART WITH 101;

-- Pipeline related END
-----------------------
-- Plan related BEGIN
CREATE TABLE plan (
    id BIGSERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    pipeline_id INTEGER REFERENCES pipeline (id),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_plan_project_id ON plan(project_id);

CREATE INDEX idx_plan_pipeline_id ON plan(pipeline_id);

ALTER SEQUENCE plan_id_seq RESTART WITH 101;

CREATE TABLE plan_check_run (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    plan_id BIGINT NOT NULL REFERENCES plan (id),
    status TEXT NOT NULL CHECK (status IN ('RUNNING', 'DONE', 'FAILED', 'CANCELED')),
    type TEXT NOT NULL CHECK (type LIKE 'bb.plan-check.%'),
    config JSONB NOT NULL DEFAULT '{}',
    result JSONB NOT NULL DEFAULT '{}',
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_plan_check_run_plan_id ON plan_check_run (plan_id);

ALTER SEQUENCE plan_check_run_id_seq RESTART WITH 101;

-- Plan related END
-----------------------
-- issue
CREATE TABLE issue (
    id SERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    plan_id BIGINT REFERENCES plan (id),
    pipeline_id INTEGER REFERENCES pipeline (id),
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'DONE', 'CANCELED')),
    type TEXT NOT NULL CHECK (type LIKE 'bb.issue.%'),
    description TEXT NOT NULL DEFAULT '',
    assignee_id INTEGER REFERENCES principal (id),
    assignee_need_attention BOOLEAN NOT NULL DEFAULT FALSE, 
    payload JSONB NOT NULL DEFAULT '{}',
    ts_vector TSVECTOR
);

CREATE INDEX idx_issue_project_id ON issue(project_id);

CREATE INDEX idx_issue_plan_id ON issue(plan_id);

CREATE INDEX idx_issue_pipeline_id ON issue(pipeline_id);

CREATE INDEX idx_issue_creator_id ON issue(creator_id);

CREATE INDEX idx_issue_assignee_id ON issue(assignee_id);

CREATE INDEX idx_issue_ts_vector ON issue USING GIN(ts_vector);

ALTER SEQUENCE issue_id_seq RESTART WITH 101;

-- stores the issue subscribers.
CREATE TABLE issue_subscriber (
    issue_id INTEGER NOT NULL REFERENCES issue (id),
    subscriber_id INTEGER NOT NULL REFERENCES principal (id),
    PRIMARY KEY (issue_id, subscriber_id)
);

CREATE INDEX idx_issue_subscriber_subscriber_id ON issue_subscriber(subscriber_id);

-- instance change history records the changes an instance and its databases.
CREATE TABLE instance_change_history (
    id BIGSERIAL PRIMARY KEY,
    status TEXT NOT NULL CONSTRAINT instance_change_history_status_check CHECK (status IN ('PENDING', 'DONE', 'FAILED')),
    version TEXT NOT NULL,
    execution_duration_ns BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_instance_change_history_unique_version ON instance_change_history (version);

ALTER SEQUENCE instance_change_history_id_seq RESTART WITH 101;

CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

CREATE INDEX idx_audit_log_payload_parent ON audit_log((payload->>'parent'));

CREATE INDEX idx_audit_log_payload_method ON audit_log((payload->>'method'));

CREATE INDEX idx_audit_log_payload_resource ON audit_log((payload->>'resource'));

CREATE INDEX idx_audit_log_payload_user ON audit_log((payload->>'user'));

ALTER SEQUENCE audit_log_id_seq RESTART WITH 101;

CREATE TABLE issue_comment (
    id BIGSERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    issue_id INTEGER NOT NULL REFERENCES issue (id),
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_issue_comment_issue_id ON issue_comment(issue_id);

ALTER SEQUENCE issue_comment_id_seq RESTART WITH 101;

CREATE TABLE query_history (
    id BIGSERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id TEXT NOT NULL, -- the project resource id
    database TEXT NOT NULL, -- the database resource name, for example, instances/{instance}/databases/{database}
    statement TEXT NOT NULL,
    type TEXT NOT NULL, -- the history type, support QUERY and EXPORT.
    payload JSONB NOT NULL DEFAULT '{}' -- saved for details, like error, duration, etc.
);

CREATE INDEX idx_query_history_creator_id_created_at_project_id ON query_history(creator_id, created_at, project_id DESC);

ALTER SEQUENCE query_history_id_seq RESTART WITH 101;

-- vcs table stores the version control provider config
CREATE TABLE vcs (
    id SERIAL PRIMARY KEY,
    resource_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('GITLAB', 'GITHUB', 'BITBUCKET', 'AZURE_DEVOPS')),
    instance_url TEXT NOT NULL CHECK ((instance_url LIKE 'http://%' OR instance_url LIKE 'https://%') AND instance_url = rtrim(instance_url, '/')),
    access_token TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_vcs_unique_resource_id ON vcs(resource_id);

ALTER SEQUENCE vcs_id_seq RESTART WITH 101;

-- vcs_connector table stores vcs connectors for a project
CREATE TABLE vcs_connector (
    id SERIAL PRIMARY KEY,
    vcs_id INTEGER NOT NULL REFERENCES vcs (id),
    project_id INTEGER NOT NULL REFERENCES project (id),
    resource_id TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_vcs_connector_unique_project_id_resource_id ON vcs_connector(project_id, resource_id);

ALTER SEQUENCE vcs_connector_id_seq RESTART WITH 101;

-- Anomaly
-- anomaly stores various anomalies found by the scanner.
-- For now, anomaly can be associated with a particular instance or database.
CREATE TABLE anomaly (
    id SERIAL PRIMARY KEY,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project TEXT NOT NULL,
    instance_id INTEGER NOT NULL REFERENCES instance (id),
    database_id INTEGER NULL REFERENCES db (id),
    type TEXT NOT NULL CHECK (type LIKE 'bb.anomaly.%'),
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_anomaly_unique_project_database_id_type ON anomaly(project, database_id, type);

ALTER SEQUENCE anomaly_id_seq RESTART WITH 101;

-- Deployment Configuration.
-- deployment_config stores deployment configurations at project level.
CREATE TABLE deployment_config (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project (id),
    name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_deployment_config_unique_project_id ON deployment_config(project_id);

ALTER SEQUENCE deployment_config_id_seq RESTART WITH 101;

-- worksheet table stores worksheets in SQL Editor.
CREATE TABLE worksheet (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    database_id INTEGER NULL REFERENCES db (id),
    name TEXT NOT NULL,
    statement TEXT NOT NULL,
    visibility TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_worksheet_creator_id_project_id ON worksheet(creator_id, project_id);

ALTER SEQUENCE worksheet_id_seq RESTART WITH 101;

-- worksheet_organizer table stores the sheet status for a principal.
CREATE TABLE worksheet_organizer (
    id SERIAL PRIMARY KEY,
    worksheet_id INTEGER NOT NULL REFERENCES worksheet (id) ON DELETE CASCADE,
    principal_id INTEGER NOT NULL REFERENCES principal (id),
    starred BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX idx_worksheet_organizer_unique_sheet_id_principal_id ON worksheet_organizer(worksheet_id, principal_id);

CREATE INDEX idx_worksheet_organizer_principal_id ON worksheet_organizer(principal_id);

-- external_approval stores approval instances of third party applications.
CREATE TABLE external_approval ( 
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL REFERENCES issue (id),
    requester_id INTEGER NOT NULL REFERENCES principal (id),
    approver_id INTEGER NOT NULL REFERENCES principal (id),
    type TEXT NOT NULL CHECK (type LIKE 'bb.plugin.app.%'),
    payload JSONB NOT NULL
);

CREATE INDEX idx_external_approval_issue_id ON external_approval(issue_id);

ALTER SEQUENCE external_approval_id_seq RESTART WITH 101;

-- risk stores the definition of a risk.
CREATE TABLE risk (
    id BIGSERIAL PRIMARY KEY,
    source TEXT NOT NULL CHECK (source LIKE 'bb.risk.%'),
    -- how risky is the risk, the higher the riskier
    level BIGINT NOT NULL,
    name TEXT NOT NULL,
    active BOOLEAN NOT NULL,
    expression JSONB NOT NULL
);

ALTER SEQUENCE risk_id_seq RESTART WITH 101;

-- slow_query stores slow query statistics for each database.
CREATE TABLE slow_query (
    id SERIAL PRIMARY KEY,
    -- In MySQL, users can query without specifying a database. In this case, instance_id is used to identify the instance.
    instance_id INTEGER NOT NULL REFERENCES instance (id),
    -- In MySQL, users can query without specifying a database. In this case, database_id is NULL.
    database_id INTEGER NULL REFERENCES db (id),
    -- It's hard to store all slow query logs, so the slow query is aggregated by day and database.
    log_date_ts INTEGER NOT NULL,
    -- It's hard to store all slow query logs, we sample the slow query log and store the part of them as details.
    slow_query_statistics JSONB NOT NULL DEFAULT '{}'
);

-- The slow query log is aggregated by day and database and we usually query the slow query log by day and database.
CREATE UNIQUE INDEX uk_slow_query_database_id_log_date_ts ON slow_query (database_id, log_date_ts);

CREATE INDEX idx_slow_query_instance_id_log_date_ts ON slow_query (instance_id, log_date_ts);

ALTER SEQUENCE slow_query_id_seq RESTART WITH 101;

CREATE TABLE db_group (
    id BIGSERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project (id),
    resource_id TEXT NOT NULL,
    placeholder TEXT NOT NULL DEFAULT '',
    expression JSONB NOT NULL DEFAULT '{}',
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_db_group_unique_project_id_resource_id ON db_group(project_id, resource_id);

CREATE UNIQUE INDEX idx_db_group_unique_project_id_placeholder ON db_group(project_id, placeholder);

ALTER SEQUENCE db_group_id_seq RESTART WITH 101;

-- changelist table stores project changelists.
CREATE TABLE changelist (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    project_id INTEGER NOT NULL REFERENCES project (id),
    name TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_changelist_project_id_name ON changelist(project_id, name);

ALTER SEQUENCE changelist_id_seq RESTART WITH 101;

CREATE TABLE export_archive (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  bytes BYTEA,
  payload JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE user_group (
  email TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  payload JSONB NOT NULL DEFAULT '{}'
);

-- review config table.
CREATE TABLE review_config (
    id TEXT NOT NULL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    name TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE revision (
    id BIGSERIAL PRIMARY KEY,
    database_id INTEGER NOT NULL REFERENCES db (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleter_id INTEGER REFERENCES principal (id),
    deleted_at TIMESTAMPTZ,
    version TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'
);

ALTER SEQUENCE revision_id_seq RESTART WITH 101;

CREATE UNIQUE INDEX IF NOT EXISTS idx_revision_unique_database_id_version_deleted_at_null ON revision (database_id, version) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_revision_database_id_version ON revision (database_id, version);

CREATE TABLE sync_history (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    database_id INTEGER NOT NULL REFERENCES db (id),
    metadata JSON NOT NULL DEFAULT '{}',
    raw_dump TEXT NOT NULL DEFAULT ''
);

ALTER SEQUENCE sync_history_id_seq RESTART WITH 101;

CREATE INDEX IF NOT EXISTS idx_sync_history_database_id_created_at ON sync_history (database_id, created_at);

CREATE TABLE changelog (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    database_id INTEGER NOT NULL REFERENCES db (id),
    status TEXT NOT NULL CONSTRAINT changelog_status_check CHECK (status IN ('PENDING', 'DONE', 'FAILED')),
    prev_sync_history_id BIGINT REFERENCES sync_history (id),
    sync_history_id BIGINT REFERENCES sync_history (id),
    payload JSONB NOT NULL DEFAULT '{}'
);

ALTER SEQUENCE changelog_id_seq RESTART WITH 101;

CREATE INDEX IF NOT EXISTS idx_changelog_database_id ON changelog (database_id);

CREATE TABLE IF NOT EXISTS release (
    id BIGSERIAL PRIMARY KEY,
    row_status row_status NOT NULL DEFAULT 'NORMAL',
    project_id INTEGER NOT NULL REFERENCES project (id),
    creator_id INTEGER NOT NULL REFERENCES principal (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload JSONB NOT NULL DEFAULT '{}'
);

ALTER SEQUENCE release_id_seq RESTART WITH 101;

CREATE INDEX idx_release_project_id ON release (project_id);


-- Default bytebase system account id is 1.
INSERT INTO principal (id, type, name, email, password_hash) VALUES (1, 'SYSTEM_BOT', 'Bytebase', 'support@bytebase.com', '');

ALTER SEQUENCE principal_id_seq RESTART WITH 101;

-- Default project.
INSERT INTO project (id, name, resource_id) VALUES (1, 'Default', 'default');

ALTER SEQUENCE project_id_seq RESTART WITH 101;

-- Create "test" and "prod" environments
INSERT INTO environment (id, name, "order", resource_id) VALUES (101, 'Test', 0, 'test');
INSERT INTO environment (id, name, "order", resource_id) VALUES (102, 'Prod', 1, 'prod');

ALTER SEQUENCE environment_id_seq RESTART WITH 103;
