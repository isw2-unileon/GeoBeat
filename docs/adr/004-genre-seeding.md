# ADR-004: Preload genres into the database

## Status

Accepted

## Date

2026-06-07

## Context

Game sessions require a known set of genres to generate challenges safely. Inserting genres dynamically during play can lead to malformed entries, duplicate values, or inconsistent application state.

## Decision

Seed the database with the allowed genre list before gameplay begins. Genres are inserted ahead of time to preserve data integrity and prevent garbage or unexpected genre values during game sessions.

## Consequences

- The backend can rely on a fixed genre catalog when generating challenges.
- Challenge generation is faster because genre data is already present.
- Initial setup requires a seeding step, but runtime stability improves.
