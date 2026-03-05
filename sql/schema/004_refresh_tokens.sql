-- +goose Up
CREATE TABLE refreshTokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID REFERENCES users ON DELETE CASCADE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

-- +goose DOWN
DROP TABLE refreshTokens;