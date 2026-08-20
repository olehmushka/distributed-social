CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS moderation_actions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    text NOT NULL,
    target_type text NOT NULL,
    target_id   text NOT NULL,
    action      text NOT NULL,
    reason      text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS moderation_actions_target_idx ON moderation_actions (target_type, target_id);
