CREATE TABLE IF NOT EXISTS analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES alerts(id),
    group_id UUID REFERENCES alert_groups(id),
    type VARCHAR(50) NOT NULL DEFAULT 'root_cause',
    summary TEXT NOT NULL DEFAULT '',
    root_cause TEXT NOT NULL DEFAULT '',
    suggestions JSONB NOT NULL DEFAULT '[]',
    severity_suggestion VARCHAR(50) DEFAULT '',
    context_snapshot JSONB DEFAULT '{}',
    llm_model VARCHAR(100) NOT NULL DEFAULT '',
    llm_tokens_used INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analyses_alert_id ON analyses(alert_id);
CREATE INDEX idx_analyses_group_id ON analyses(group_id);
CREATE INDEX idx_analyses_created_at ON analyses(created_at DESC);
