## Review & Suggestions for Improvements

Great work on Phase 1! The foundation is solid, and the separation of concerns is very clean. I've reviewed the code and here are a few suggestions for improvement and optimization:

### 1. WebSocket Buffer Lifecycle (Concurrent Connections)
In `internal/websocket/hub.go`, `h.buffer.FlushAndClear(rideID)` is called in the `defer` block of `HandleWS`. If a user reconnects quickly before the old connection fully closes, or connects from two devices with the same `rideID`, the older connection's closure will clear the active buffer and wipe out pending points. 
**Optimization:** Consider using a reference count (tracking active connections per `rideID`) or adding a session ID so that you only `FlushAndClear` when the *last* connection for a ride closes.

### 2. Server Startup Error Handling
In `cmd/server/main.go`, the HTTP server is started in a goroutine that calls `log.Fatalf` if `ListenAndServe()` fails.
**Improvement:** Instead of immediately killing the process with `log.Fatalf` inside the goroutine, send the error to an error channel. You can then `select` on both the OS signal channel and the error channel in the main thread. This ensures that if the server fails to bind (e.g., port in use), the app can still shut down the DB, Redis, and Cron gracefully.

### 3. Timer Reset Overhead in GPSBuffer
In `internal/websocket/buffer.go`, `buf.timer = time.AfterFunc(...)` is used to reset the flush timer on every GPS point received. 
**Optimization:** `time.AfterFunc` spawns a new background goroutine each time it fires, and stopping/recreating timers frequently can be overhead-heavy. An alternative pattern is running a single background goroutine (or ticker) per active ride that periodically checks if flushing is needed based on a "last updated" timestamp, or relying solely on batch limits combined with a global flush ticker.

### 4. Hardcoded Configurations
- Database pool `MaxConns=25` and `MinConns=5` in `main.go`.
- The timezone `mustLoadLocation("Asia/Jakarta")` in `internal/jobs/leaderboard.go`.
**Improvement:** Moving these to `config.go` and `.env` will make the application more flexible across different environments.

---

## Implementation Stages (Low-Level Summary)
For tracking purposes, here is a summary of the implementation stages achieved in this PR:
1. **Foundation & Config:** Project scaffolding, Go module initialization, and `Config` struct with strict environment parsing and defaults.
2. **Database & Cache Bootstrapping:** `pgxpool` and `redis` client initialization in `main.go` with startup health checks (Pings) and graceful shutdown handling.
3. **Core Middleware:** Implementation of JWT-based Auth (handling token extraction and specific expired/invalid states) and standard CORS configurations.
4. **WebSocket Infrastructure:** Creation of the `GPSBuffer` for batching high-frequency GPS points to minimize DB inserts, paired with a WebSocket `Hub` that validates connections using short-lived tokens stored in Redis.
5. **Routing & Handlers:** Setup of the Gin router defining the public/protected API boundary and stubs for all Phase 2 business logic handlers.
6. **Background Jobs:** Integration of `robfig/cron` for the weekly leaderboard generation, complete with lifecycle management (Start/Stop).
