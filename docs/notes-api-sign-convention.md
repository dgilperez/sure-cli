# Sure API v1: amount sign convention (important)

Sure API v1 may return `amount` with an **accounting-style sign convention** that looks inverted compared to the UI:

- `classification=income` can come with a **negative** `amount` string
- `classification=expense` can come with a **positive** `amount` string

This is reported as intentional upstream behavior (see issue #893 discussion).

## Practical rule for agents / automation

When computing signed values, treat `classification` as ground truth:

- income => positive
- expense => negative

In `sure-cli` **insights** we normalize sign this way so heuristics remain intuitive.

## Recommendation for upstream API

For agent-first/deterministic consumption, the API should expose numeric fields:

- `amount_cents` (absolute)
- `signed_amount_cents` (income positive, expense negative)

and keep `amount` as a human-facing formatted string if desired.
