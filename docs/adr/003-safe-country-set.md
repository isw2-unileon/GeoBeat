# ADR-003: Restrict playable countries to a safe subset

## Status

Accepted

## Date

2026-06-07

## Context

Not all countries have enough accessible or consistent music metadata to support gameplay. Using an unrestricted country list risks generating challenges with missing songs or incomplete genre coverage.

## Decision

Limit playable countries to a curated safe subset that has reliable music data availability. This ensures that every challenge can be generated with valid songs and reduces the risk of runtime failures caused by missing content.

## Consequences

- Game generation becomes more stable and predictable.
- Some countries are intentionally excluded even if they are geographically valid.
- The playable map is smaller, but user experience is more reliable.
