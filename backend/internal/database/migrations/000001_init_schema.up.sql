CREATE TABLE daily_challenges (
    id SERIAL PRIMARY KEY,
    target_country VARCHAR(255) NOT NULL,
    target_genre VARCHAR(255) NOT NULL,
    hint_songs TEXT[] NOT NULL,
    play_date DATE UNIQUE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE daily_sessions (
    user_id UUID NOT NULL,
    challenge_id INT NOT NULL REFERENCES daily_challenges(id),
    attempts_used INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    
    PRIMARY KEY (user_id, challenge_id) 
);

CREATE TABLE genres (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    normalized_name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE tracks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    artist VARCHAR(255) NOT NULL,
    genres TEXT[] NOT NULL
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    user_name VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),  -- Nullable for OAuth users
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255),    -- Nullable for Email users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_provider CHECK (provider IN ('email', 'google')),

    CONSTRAINT check_auth_data CHECK (
        (provider = 'email' AND password_hash IS NOT NULL) OR
        (provider != 'email' AND provider_id IS NOT NULL)
    )
);