# Reusable Cache (Business Value)

Provides a small, concurrency‑safe, time‑to‑live key/value cache used by context, pricing, or LLM layers. Benefits:

- Reduces repeated external API calls (cost + latency savings).
- Consistent diagnostics (hit/miss counts) across features.
- Pluggable: can be swapped out for a distributed cache in future.

Designed for modest in‑process workloads; not a replacement for a large scale distributed store.
