# File Access (Business Value)

Provides a lightweight, swappable file system layer so higher‑level features (reports, context packs, exports) can store and retrieve artifacts without re‑implementing OS concerns. This enables:

- Seamless redirection to in‑memory or test file systems for faster automation.
- Centralised, auditable persistence paths (one base directory decision).
- Future remote backends (object storage, encrypted vault) without changing business logic.

End users indirectly benefit through more reliable, atomic writes of reports and exports.
