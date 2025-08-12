# Persistence Manager (Business Value)

Unified, atomic persistence layer for reports, context packs, and exports:

- Single entry point to write JSON, Markdown, PDF.
- Ensures directories exist & writes are atomic (temp + rename) reducing corruption risk.
- Abstracts file system & PDF converter so tests run without external tools.
- Eases future move to encrypted or remote storage backends.

Users benefit from reliable generation of artifacts they can archive or share.
