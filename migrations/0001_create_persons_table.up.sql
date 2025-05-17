CREATE TABLE IF NOT EXISTS persons (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    patronymic TEXT,
    age INT,
    gender TEXT,
    nationality TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);