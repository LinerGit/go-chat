CREATE TABLE IF NOT EXISTS messages (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    NOT NULL,
    username   TEXT      NOT NULL,
    content    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_created_at ON messages (created_at DESC);
CREATE INDEX idx_messages_user_id    ON messages (user_id);
