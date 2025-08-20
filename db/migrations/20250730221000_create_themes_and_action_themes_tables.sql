-- migrate:up
CREATE TABLE theme (
    id BYTEA PRIMARY KEY,
    person_id BYTEA NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    text TEXT NOT NULL CHECK (LENGTH(TRIM(BOTH FROM text)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_theme_person_id ON theme(person_id);
CREATE INDEX idx_theme_created_at ON theme(created_at);

CREATE TRIGGER update_theme_updated_at
    BEFORE UPDATE ON theme
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE action_theme (
    action_id BYTEA NOT NULL REFERENCES action(id) ON DELETE CASCADE,
    theme_id BYTEA NOT NULL REFERENCES theme(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (action_id, theme_id)
);

CREATE INDEX idx_action_theme_action_id ON action_theme(action_id);
CREATE INDEX idx_action_theme_theme_id ON action_theme(theme_id);
CREATE INDEX idx_action_theme_created_at ON action_theme(created_at);

CREATE TRIGGER update_action_theme_updated_at
    BEFORE UPDATE ON action_theme
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- migrate:down
DROP TRIGGER IF EXISTS update_action_theme_updated_at ON action_theme;
DROP INDEX IF EXISTS idx_action_theme_created_at;
DROP INDEX IF EXISTS idx_action_theme_theme_id;
DROP INDEX IF EXISTS idx_action_theme_action_id;
DROP TABLE IF EXISTS action_theme;

DROP TRIGGER IF EXISTS update_theme_updated_at ON theme;
DROP INDEX IF EXISTS idx_theme_created_at;
DROP INDEX IF EXISTS idx_theme_person_id;
DROP TABLE IF EXISTS theme;
