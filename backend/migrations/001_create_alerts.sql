CREATE TABLE IF NOT EXISTS alert_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL DEFAULT 'default',
    status VARCHAR(50) NOT NULL DEFAULT 'firing',
    alert_count INT NOT NULL DEFAULT 0,
    root_cause_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_groups_namespace ON alert_groups(namespace);
CREATE INDEX idx_alert_groups_status ON alert_groups(status);

CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fingerprint VARCHAR(64) NOT NULL,
    group_id UUID REFERENCES alert_groups(id),
    status VARCHAR(50) NOT NULL DEFAULT 'firing',
    severity VARCHAR(50) NOT NULL DEFAULT 'warning',
    adjusted_severity VARCHAR(50),
    name VARCHAR(255) NOT NULL DEFAULT '',
    namespace VARCHAR(255) NOT NULL DEFAULT '',
    pod VARCHAR(255) DEFAULT '',
    node VARCHAR(255) DEFAULT '',
    labels JSONB NOT NULL DEFAULT '{}',
    annotations JSONB NOT NULL DEFAULT '{}',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_by VARCHAR(255) DEFAULT '',
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_alerts_fingerprint ON alerts(fingerprint);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_namespace ON alerts(namespace);
CREATE INDEX idx_alerts_group_id ON alerts(group_id);
CREATE INDEX idx_alerts_created_at ON alerts(created_at DESC);

CREATE TABLE IF NOT EXISTS silence_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    matchers JSONB NOT NULL DEFAULT '[]',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_by VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
