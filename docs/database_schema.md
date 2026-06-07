# Data Models & Infrastructure

GeoBeat utilizes a relational database model hosted on **PostgreSQL** (Supabase for production, Docker `tmpfs` for testing).

## 🗄️ Core Entities

### 1. `users`

Stores authentication and profile data.

- `id` (UUIDv4, Primary Key)
- `email` (String, Unique)
- `user_name` (String, Unique)
- `password_hash` (String, nullable for OAuth users)
- `provider` (Enum: `email`, `google`)
- `provider_id` (String, nullable for email users)
- `created_at` (Timestamp with time zone)
- `updated_at` (Timestamp with time zone)

The `users` table enforces provider-specific authentication rules:

- email users must have `password_hash` set
- OAuth users must have `provider_id` set

### 2. `challenges`

Defines the problem type that daily and inverse challenges share.

- `id` (Serial, Primary Key)
- `challenge_type` (`daily` or `inverse`)

### 3. `daily_challenges`

Stores the playable daily challenge metadata.

- `id` (INT, Foreign Key to `challenges.id`)
- `target_country` (String)
- `target_genre` (String)
- `hint_songs` (Text array)
- `play_date` (Date, Unique, defaults to current date)

### 4. `inverse_challenges`

Stores the inverse challenge metadata.

- `id` (INT, Foreign Key to `challenges.id`)
- `given_song_name` (String)
- `target_country` (String)
- `play_date` (Date, Unique, defaults to current date)

### 5. `daily_sessions`

Tracks per-user daily game state.

- `user_id` (UUID, Foreign Key to `users.id`)
- `challenge_id` (INT, Foreign Key to `challenges.id`)
- `attempts_used` (Integer, defaults to 0)
- `status` (String)
- `updated_at` (Timestamp with time zone)
- Primary key: `(user_id, challenge_id)`

### 6. `genres`

Stores the allowed genre names used by the game.

- `id` (Serial, Primary Key)
- `name` (String, Unique)
- `normalized_name` (String, Unique)

### 7. `tracks`

Stores song metadata used by challenge generation and matching.

- `id` (Serial, Primary Key)
- `name` (String, Unique)
- `artist` (String)
- `genres` (Text array)

### 8. `refresh_tokens`

Manages refresh tokens for authenticated users.

- `id` (Serial, Primary Key)
- `user_id` (UUID, Foreign Key to `users.id`)
- `hash` (String, Unique)
- `created_at` (Timestamp with time zone)
- `expires_at` (Timestamp with time zone)

### 9. `timetrial_challenges`

Stores time trial challenge definitions.

- `id` (Serial, Primary Key)
- `target_countries` (Text array)
- `target_genres` (Text array)
- `play_date` (Date, Unique, defaults to current date)
- Constraint: `target_countries` and `target_genres` must have the same array length

### 10. `timetrial_sessions`

Tracks in-progress and completed timetrial games.

- `user_id` (UUID, Foreign Key to `users.id`)
- `challenge_id` (INT, Foreign Key to `timetrial_challenges.id`)
- `current_index` (Integer, defaults to 0)
- `start_time` (Timestamp with time zone)
- `end_time` (Timestamp with time zone, nullable)
- `duration` (BigInt, nullable)
- `status` (String, values: `playing`, `completed`)
- Primary key: `(user_id, challenge_id)`
- Leaderboard index: `idx_timetrial_leaderboard` on `(challenge_id, duration ASC)` for completed games

---

## 🛠️ Infrastructure Decisions

### UUIDs and Serial IDs

The repository uses UUIDv4 for user identifiers and user-linked session references, while challenge-related tables and token tables use serial integer primary keys where appropriate.

### Time Duration Handling

The backend stores timetrial `duration` as a PostgreSQL `BIGINT`. In Go, durations are converted using `Duration.Milliseconds()` when reading from or writing to the database, preserving compatibility with frontend timing logic.
