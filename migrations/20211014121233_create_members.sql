-- +goose Up
-- +goose StatementBegin
CREATE TABLE members (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL, 
    member_id VARCHAR PRIMARY KEY, 
    name VARCHAR NOT NULL, 
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    room_id VARCHAR NOT NULL REFERENCES rooms(room_id)
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON members
FOR EACH STATEMENT
EXECUTE PROCEDURE trigger_set_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER set_updated_at ON members;
DROP TABLE members;
-- +goose StatementEnd
