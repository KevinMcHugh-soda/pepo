-- migrate:up

-- First, create helper functions for XID conversion
-- XID uses Crockford's Base32 encoding (0123456789ABCDEFGHJKMNPQRSTVWXYZ)
-- An XID is 20 characters representing 12 bytes (96 bits)

-- Function to convert XID string to bytea
-- XID uses base32 encoding with alphabet '0123456789abcdefghijklmnopqrstuv'
CREATE OR REPLACE FUNCTION x2b(xid_str TEXT) RETURNS BYTEA AS $$
DECLARE
    alphabet TEXT := '0123456789abcdefghijklmnopqrstuv';
    result_hex TEXT := '';
    i INTEGER;
    char_val INTEGER;
    bits BIGINT := 0;
    bit_count INTEGER := 0;
    byte_val INTEGER;
BEGIN
    -- Validate input length
    IF LENGTH(xid_str) != 20 THEN
        RAISE EXCEPTION 'XID must be exactly 20 characters long';
    END IF;

    -- Process each character of the XID
    FOR i IN 1..20 LOOP
        -- Get the position of the character in the alphabet (0-based)
        char_val := POSITION(SUBSTRING(xid_str FROM i FOR 1) IN alphabet) - 1;
        IF char_val < 0 THEN
            RAISE EXCEPTION 'Invalid character in XID: %', SUBSTRING(xid_str FROM i FOR 1);
        END IF;

        -- Accumulate 5 bits from this character
        bits := (bits << 5) | char_val;
        bit_count := bit_count + 5;

        -- Extract complete bytes (8 bits each)
        WHILE bit_count >= 8 LOOP
            byte_val := (bits >> (bit_count - 8)) & 255;
            result_hex := result_hex || LPAD(TO_HEX(byte_val), 2, '0');
            bit_count := bit_count - 8;
        END LOOP;
    END LOOP;

    -- The result should be exactly 12 bytes (24 hex characters)
    IF LENGTH(result_hex) != 24 THEN
        RAISE EXCEPTION 'XID conversion error: expected 24 hex chars, got %', LENGTH(result_hex);
    END IF;

    RETURN DECODE(result_hex, 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- Function to convert bytea to XID string
CREATE OR REPLACE FUNCTION b2x(xid_bytes BYTEA) RETURNS TEXT AS $$
DECLARE
    alphabet TEXT := '0123456789abcdefghijklmnopqrstuv';
    result TEXT := '';
    i INTEGER;
    bits BIGINT := 0;
    bit_count INTEGER := 0;
    byte_val INTEGER;
    char_index INTEGER;
BEGIN
    -- Validate input length (XID should be 12 bytes)
    IF LENGTH(xid_bytes) != 12 THEN
        RAISE EXCEPTION 'XID bytea must be exactly 12 bytes long';
    END IF;

    -- Process each byte from the bytea
    FOR i IN 0..11 LOOP
        byte_val := GET_BYTE(xid_bytes, i);
        -- Add 8 bits to the bit buffer
        bits := (bits << 8) | byte_val;
        bit_count := bit_count + 8;

        -- Extract 5-bit characters while we have enough bits
        WHILE bit_count >= 5 LOOP
            char_index := (bits >> (bit_count - 5)) & 31;
            result := result || SUBSTRING(alphabet FROM char_index + 1 FOR 1);
            bit_count := bit_count - 5;
        END LOOP;
    END LOOP;

    -- Handle remaining bits (the last character uses fewer than 5 bits)
    -- XID: 20 chars * 5 bits = 100 bits, but we only have 96 bits (12 bytes)
    -- So the last character represents the remaining 4 bits, left-padded
    IF bit_count > 0 THEN
        char_index := (bits << (5 - bit_count)) & 31;
        result := result || SUBSTRING(alphabet FROM char_index + 1 FOR 1);
    END IF;

    -- We should have exactly 20 characters
    IF LENGTH(result) != 20 THEN
        RAISE EXCEPTION 'XID conversion error: expected 20 chars, got %', LENGTH(result);
    END IF;

    RETURN result;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- Create the new person table with bytea id
CREATE TABLE person (
    id BYTEA PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes on the new table
CREATE INDEX idx_person_name ON person(name);
CREATE INDEX idx_person_created_at ON person(created_at);

-- Create trigger function for updated_at (reuse existing function)
CREATE TRIGGER update_person_updated_at
    BEFORE UPDATE ON person
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Migrate data from persons to person if it exists
INSERT INTO person (id, name, created_at, updated_at)
SELECT x2b(id), name, created_at, updated_at
FROM persons
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'persons');

-- Drop the old persons table and its indexes/triggers
DROP TRIGGER IF EXISTS update_persons_updated_at ON persons;
DROP INDEX IF EXISTS idx_persons_created_at;
DROP INDEX IF EXISTS idx_persons_name;
DROP TABLE IF EXISTS persons;

-- migrate:down

-- Recreate the old persons table
CREATE TABLE persons (
    id VARCHAR(20) PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Recreate indexes on persons
CREATE INDEX idx_persons_name ON persons(name);
CREATE INDEX idx_persons_created_at ON persons(created_at);

-- Recreate trigger
CREATE TRIGGER update_persons_updated_at
    BEFORE UPDATE ON persons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Migrate data back from person to persons if it exists
INSERT INTO persons (id, name, created_at, updated_at)
SELECT b2x(id), name, created_at, updated_at
FROM person
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'person');

-- Drop the new person table and its indexes/triggers
DROP TRIGGER IF EXISTS update_person_updated_at ON person;
DROP INDEX IF EXISTS idx_person_created_at;
DROP INDEX IF EXISTS idx_person_name;
DROP TABLE IF EXISTS person;

-- Drop helper functions
DROP FUNCTION IF EXISTS x2b(TEXT);
DROP FUNCTION IF EXISTS b2x(BYTEA);
