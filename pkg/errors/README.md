# Errors

Clear messages, no surprises.

## What you’ll see

- Human‑readable explanations when something goes wrong (missing file, bad number, unknown ticker).
- Consistent wording across commands so you can quickly fix the issue and retry.

## Typical situations

- Your portfolio file can’t be found or parsed.
- A position uses an unsupported asset type.
- A value is negative or out of range.

## How to recover

- Re‑check the file path and YAML formatting.
- Use the sample at `data/positions.sample.yaml` as a template.
- Adjust the offending field; the message will tell you which one.

## Philosophy

- Fail fast and clearly.
- Prefer guiding the user to a fix over cryptic stack traces.
