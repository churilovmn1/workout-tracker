CREATE TABLE IF NOT EXISTS template_exercises (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    sets INTEGER NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    weight_kg NUMERIC(6,2) NOT NULL DEFAULT 0
);

CREATE INDEX idx_template_exercises_template_id ON template_exercises(template_id);
