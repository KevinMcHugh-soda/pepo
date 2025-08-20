-- migrate:up
CREATE TABLE conversation (
    id BYTEA PRIMARY KEY,
    description TEXT NOT NULL CHECK (LENGTH(TRIM(BOTH FROM description)) > 0),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversation_occurred_at ON conversation(occurred_at DESC);
CREATE INDEX idx_conversation_created_at ON conversation(created_at);

CREATE TRIGGER update_conversation_updated_at
    BEFORE UPDATE ON conversation
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE action_conversation (
    action_id BYTEA NOT NULL REFERENCES action(id) ON DELETE CASCADE,
    conversation_id BYTEA NOT NULL REFERENCES conversation(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (action_id, conversation_id)
);

CREATE INDEX idx_action_conversation_action_id ON action_conversation(action_id);
CREATE INDEX idx_action_conversation_conversation_id ON action_conversation(conversation_id);
CREATE INDEX idx_action_conversation_created_at ON action_conversation(created_at);

CREATE TRIGGER update_action_conversation_updated_at
    BEFORE UPDATE ON action_conversation
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE conversation_theme (
    conversation_id BYTEA NOT NULL REFERENCES conversation(id) ON DELETE CASCADE,
    theme_id BYTEA NOT NULL REFERENCES theme(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, theme_id)
);

CREATE INDEX idx_conversation_theme_conversation_id ON conversation_theme(conversation_id);
CREATE INDEX idx_conversation_theme_theme_id ON conversation_theme(theme_id);
CREATE INDEX idx_conversation_theme_created_at ON conversation_theme(created_at);

CREATE TRIGGER update_conversation_theme_updated_at
    BEFORE UPDATE ON conversation_theme
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- migrate:down
DROP TRIGGER IF EXISTS update_conversation_theme_updated_at ON conversation_theme;
DROP INDEX IF EXISTS idx_conversation_theme_created_at;
DROP INDEX IF EXISTS idx_conversation_theme_theme_id;
DROP INDEX IF EXISTS idx_conversation_theme_conversation_id;
DROP TABLE IF EXISTS conversation_theme;

DROP TRIGGER IF EXISTS update_action_conversation_updated_at ON action_conversation;
DROP INDEX IF EXISTS idx_action_conversation_created_at;
DROP INDEX IF EXISTS idx_action_conversation_conversation_id;
DROP INDEX IF EXISTS idx_action_conversation_action_id;
DROP TABLE IF EXISTS action_conversation;

DROP TRIGGER IF EXISTS update_conversation_updated_at ON conversation;
DROP INDEX IF EXISTS idx_conversation_created_at;
DROP INDEX IF EXISTS idx_conversation_occurred_at;
DROP TABLE IF EXISTS conversation;
