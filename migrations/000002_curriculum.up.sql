CREATE TABLE curriculum_phases (
    id SERIAL PRIMARY KEY,
    phase_number INT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    description TEXT
);

CREATE TABLE curriculum_weeks (
    id SERIAL PRIMARY KEY,
    phase_id INT NOT NULL REFERENCES curriculum_phases(id),
    week_number INT NOT NULL,
    title TEXT NOT NULL,
    theme TEXT NOT NULL,
    UNIQUE(phase_id, week_number)
);

CREATE TABLE curriculum_days (
    id SERIAL PRIMARY KEY,
    week_id INT NOT NULL REFERENCES curriculum_weeks(id),
    day_name TEXT NOT NULL,
    day_order INT NOT NULL,
    focus TEXT NOT NULL
);

CREATE TABLE curriculum_tasks (
    id SERIAL PRIMARY KEY,
    day_id INT NOT NULL REFERENCES curriculum_days(id),
    task_order INT NOT NULL,
    duration_min INT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    detail TEXT
);
