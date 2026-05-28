CREATE TABLE schedule (
    id               SERIAL PRIMARY KEY,
    trainer_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id        INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    scheduled_at     TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 60,
    status           TEXT NOT NULL DEFAULT 'planned'
                     CHECK (status IN ('planned', 'completed', 'cancelled')),
    notes            TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON schedule(trainer_id, scheduled_at);
