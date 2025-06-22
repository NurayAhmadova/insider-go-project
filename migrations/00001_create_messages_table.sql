-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE FUNCTION set_current_timestamp_updated_at()
    RETURNS TRIGGER AS $$
DECLARE
    _new record;
BEGIN
    _new := NEW;
    _new."updated_at" = now();
    RETURN _new;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS messages (
                                        id          UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        msisdn      TEXT NOT NULL CHECK(char_length(msisdn) <= 15),
                                        content     TEXT NOT NULL CHECK(char_length(content) <= 160),
                                        sent        BOOLEAN NOT NULL DEFAULT FALSE,
                                        created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
                                        updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW
EXECUTE PROCEDURE set_current_timestamp_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
