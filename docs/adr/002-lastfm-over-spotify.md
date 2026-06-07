# ADR-002: Use Last.fm instead of Spotify for genre-country mapping

## Status

Accepted

## Date

2026-06-07

## Context

GeoBeat needs a music data provider that can reliably map musical genres to specific countries. The Spotify API does not provide a native way to combine country and genre popularity data, which is essential for the gameplay experience.

## Decision

Use Last.fm as the primary music data source for genre-country relationships. Last.fm supports the genre geography and tagging features that GeoBeat requires to build challenges based on regional genre prominence.

## Consequences

- We can generate meaningful game content around genres tied to geographic regions.
- Dependency on Spotify is removed, avoiding brittle or incomplete data flows.
- The backend must integrate with Last.fm-specific endpoints and data transformations.
