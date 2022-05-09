# todo

-

# later

- add protocol indexer id to logs
- makefile should not contain secret, ye?
- context drilling
- proper db design
- tests
- prettier for sql
- db reset/setup/seed
- rename protocol_indexer table to just indexer?
- move type field into initial migration

# Simple Worker

- Fetch indexer specs
- Loop through indexer specs (per type)
  - Fetch blocks (per indexer spec)
  - Index blocks (per indexer spec)
  - Store users (per indexer spec)

# Simple API

- api/<address>/protocols
- List protocols -> list suggested triggers
