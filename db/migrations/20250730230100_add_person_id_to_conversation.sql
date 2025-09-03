-- migrate:up
ALTER TABLE conversation
    ADD COLUMN person_id BYTEA NOT NULL REFERENCES person(id) ON DELETE CASCADE;

CREATE INDEX idx_conversation_person_id ON conversation(person_id);

-- migrate:down
DROP INDEX IF EXISTS idx_conversation_person_id;
ALTER TABLE conversation DROP COLUMN IF EXISTS person_id;
