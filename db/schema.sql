SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: b2x(bytea); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.b2x(xid_bytes bytea) RETURNS text
    LANGUAGE plpgsql IMMUTABLE STRICT
    AS $$
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
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: x2b(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.x2b(xid_str text) RETURNS bytea
    LANGUAGE plpgsql IMMUTABLE STRICT
    AS $$
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
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: person; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.person (
    id bytea NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT person_name_check CHECK ((length(TRIM(BOTH FROM name)) > 0))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_person_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_person_created_at ON public.person USING btree (created_at);


--
-- Name: idx_person_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_person_name ON public.person USING btree (name);


--
-- Name: person update_person_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_person_updated_at BEFORE UPDATE ON public.person FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20250730100649'),
    ('20250730152732');
