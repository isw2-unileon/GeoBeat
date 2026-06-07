# Data Models & Infrastructure

GeoBeat utilizes a relational database model hosted on **PostgreSQL** (Supabase for production, Docker `tmpfs` for testing).

## 🗄️ Core Entities

### 1. `users`
Stores authentication and profile data.
* `id` (UUIDv4, Primary Key)
* `email` (String, Unique)
* `oauth_provider` (String - e.g., 'google')
* `created_at` (Timestamp)

### 2. `genres`
The foundational dataset mapping musical genres to geographical locations.
* `id` (UUIDv4)
* `country_iso` (String, length 3)
* `name` (String, Last.fm tag name)
* `weight` (Integer, popularity score)

### 3. `timetrial_sessions`
Tracks active and completed rapid-fire games.
* `id` (UUIDv4)
* `user_id` (UUIDv4, Foreign Key)
* `start_time` (Timestampz)
* `end_time` (Timestampz)
* `duration` (BigInt, stored strictly in **milliseconds**)
* `status` (Enum: 'playing', 'completed', 'abandoned')

---

## 🛠️ Infrastructure Decisions

### UUIDv4 over Auto-Incrementing IDs
We use UUIDv4 for all primary keys instead of sequential integers (1, 2, 3...). This is a crucial security decision to prevent **Insecure Direct Object Reference (IDOR)** attacks. Malicious users cannot scrape our database or guess the URL of another user's game session by simply incrementing an ID parameter.

### Time Duration Handling
In Go, `time.Duration` represents nanoseconds as an `int64`. To maintain precision and compatibility with frontend JavaScript (which operates natively in milliseconds via `Date.now()`), all backend durations are explicitly converted using `duration.Milliseconds()` before being persisted to the database.
