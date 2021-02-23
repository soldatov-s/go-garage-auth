-- +goose Up

CREATE TABLE IF NOT EXISTS production.token (
    signature character varying(255) PRIMARY KEY,
    subject character varying(255),
    meta jsonb,
    expired_at timestamp with time zone
);

-- +goose Down
DROP TABLE production.token;