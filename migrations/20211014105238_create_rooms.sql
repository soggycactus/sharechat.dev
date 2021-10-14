-- +goose Up
-- +goose StatementBegin
CREATE TABLE rooms (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL, 
    room_id VARCHAR PRIMARY KEY,
    room_name VARCHAR NOT NULL
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON rooms
FOR EACH STATEMENT
EXECUTE PROCEDURE trigger_set_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER set_updated_at ON rooms;
DROP TABLE rooms;
-- +goose StatementEnd
