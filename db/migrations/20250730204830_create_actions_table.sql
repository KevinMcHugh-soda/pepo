-- migrate:up
-- Create valence enum type
CREATE TYPE valence_type AS ENUM ('positive', 'negative');

-- Create action table
CREATE TABLE action (
    id BYTEA PRIMARY KEY,
    person_id BYTEA NOT NULL REFERENCES person(id) ON DELETE CASCADE,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    description TEXT NOT NULL,
    "references" TEXT,
    valence valence_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT action_description_check CHECK (LENGTH(TRIM(BOTH FROM description)) > 0)
);

-- Create indexes for better performance
CREATE INDEX idx_action_person_id ON action(person_id);
CREATE INDEX idx_action_occurred_at ON action(occurred_at DESC);
CREATE INDEX idx_action_valence ON action(valence);
CREATE INDEX idx_action_created_at ON action(created_at);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_action_updated_at
    BEFORE UPDATE ON action
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- migrate:down
-- Drop the action table and related objects
DROP TRIGGER IF EXISTS update_action_updated_at ON action;
DROP INDEX IF EXISTS idx_action_created_at;
DROP INDEX IF EXISTS idx_action_valence;
DROP INDEX IF EXISTS idx_action_occurred_at;
DROP INDEX IF EXISTS idx_action_person_id;
DROP TABLE IF EXISTS action;
DROP TYPE IF EXISTS valence_type;
