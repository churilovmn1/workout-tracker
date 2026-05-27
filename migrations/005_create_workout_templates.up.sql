CREATE TABLE IF NOT EXISTS workout_templates (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_workout_templates_user_id ON workout_templates(user_id);
