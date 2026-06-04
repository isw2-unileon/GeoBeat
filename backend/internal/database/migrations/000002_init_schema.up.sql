CREATE TABLE timeattack_challenges (
    id SERIAL PRIMARY KEY,
    target_countries TEXT[] NOT NULL,
    target_genres TEXT[] NOT NULL,
    play_date DATE UNIQUE NOT NULL DEFAULT CURRENT_DATE,

    CONSTRAINT arrays_length_match CHECK (array_length(target_countries, 1) = array_length(target_genres, 1))
);

CREATE TABLE timeattack_sessions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id INT NOT NULL REFERENCES timeattack_challenges(id) ON DELETE CASCADE,
    current_index INT NOT NULL DEFAULT 0,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration BIGINT, 
    status VARCHAR(50) NOT NULL,

    PRIMARY KEY (user_id, challenge_id),
    CONSTRAINT valid_status CHECK (status IN ('playing', 'completed'))
);

CREATE INDEX idx_timeattack_leaderboard 
ON timeattack_sessions (challenge_id, duration ASC) 
WHERE status = 'completed';