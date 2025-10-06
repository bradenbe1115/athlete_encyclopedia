CREATE TABLE IF NOT EXISTS athletes (
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    full_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS leagues (
    id serial PRIMARY KEY,
    league_name TEXT NOT NULL UNIQUE,
    league_sport TEXT NOT NULL,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
    id serial PRIMARY KEY,
    team_name TEXT NOT NULL UNIQUE,
    team_location_anme TEXT NOT NULL,
    team_mascot_name TEXT NOT NULL,
    league_id INTEGER REFERENCES leagues(id)
);

CREATE TABLE IF NOT EXISTS seasons (
    id serial PRIMARY KEY,
    season TEXT,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS athletes_years (
    id serial PRIMARY KEY,
    athlete_id INTEGER REFERENCES athletes(id),
    team_id INTEGER REFERENCES teams(id),
    season_id INTEGER REFERENCES seasons(id),
    created_at TIMESTAMP
);