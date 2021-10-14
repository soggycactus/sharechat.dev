-- +goose Up
-- +goose StatementBegin
CREATE TYPE message_type AS ENUM ('chat', 'joined', 'left');

CREATE TABLE messages (
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL, 
    message_id VARCHAR PRIMARY KEY, 
    type message_type NOT NULL, 
    message VARCHAR NOT NULL, 
    sent TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    room_id VARCHAR NOT NULL REFERENCES rooms(room_id),
    member_id VARCHAR NOT NULL REFERENCES members(member_id)
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON messages
FOR EACH STATEMENT
EXECUTE PROCEDURE trigger_set_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER set_updated_at ON messages;
DROP TABLE messages;
DROP TYPE message_type; 
-- +goose StatementEnd