CREATE TABLE IF NOT EXISTS k8s_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uid VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'Normal',
    reason VARCHAR(255) NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    namespace VARCHAR(255) NOT NULL DEFAULT '',
    involved_object JSONB NOT NULL DEFAULT '{}',
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_k8s_events_uid ON k8s_events(uid);
CREATE INDEX idx_k8s_events_namespace ON k8s_events(namespace);
CREATE INDEX idx_k8s_events_type ON k8s_events(type);
CREATE INDEX idx_k8s_events_last_seen ON k8s_events(last_seen DESC);
