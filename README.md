# aquifer

> A connection pool that stores and supplies connections on demand.

`aquifer` is a generic, production-grade connection pool written in Go, built from scratch.

`aquifer` handles:

- Pre warming a minimum number of connections on startup
- Lending connections to callers and taking them back
- Creating new connections on demand (up to a configurable max)
- Blocking callers gracefully when the pool is at capacity
- Timing out callers via `context.Context`
- Evicting idle connections that have expired
- Graceful shutdown — waiting for in-flight connections before closing
