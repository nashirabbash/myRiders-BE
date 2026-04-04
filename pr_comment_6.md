## PR Review: Phase 3 - Authentication and Users

Great work implementing the JWT logic and separating the concerns nicely into handlers and services. The code is clean and properly masks sensitive fields from the JSON DTOs.

However, I've identified a critical bug in the profile update flow, a race condition in the registration process, and a few opportunities for optimization. Please address the following findings:

### 🔍 Findings & Optimizations

1. **Critical: `UpdateMe` Partial Update Deletes `display_name`**
   - **Reference:** `internal/handler/users.go` (`UpdateMe`)
   - **Issue:** The `UpdateProfileRequest` correctly makes `DisplayName` optional. However, `sqlc` generated `DisplayName` as a required `string` because the DB column is `TEXT NOT NULL`. If a user only updates their `avatar_url` and omits `display_name`, `derefString(req.DisplayName)` passes `""` to the database. The `COALESCE` in the query won't catch this (since `""` is not `NULL`), and it will completely erase the user's display name.
   - **Suggestion:** Modify the SQL query in `internal/db/queries/auth.sql` to use `sqlc.narg()` so `sqlc` generates a nullable `pgtype.Text`:
     ```sql
     SET display_name = COALESCE(sqlc.narg(display_name), display_name), ...
     ```
     Alternatively, fetch the user profile first in the handler and use the existing value if `req.DisplayName` is nil.

2. **High: Race Condition & Extra DB Trips in `Register`**
   - **Reference:** `internal/service/auth.go` (`Register`)
   - **Issue:** The `Register` function checks if the email or username exists via two separate `SELECT` queries before running `INSERT`. This introduces a TOCTOU (Time-Of-Check to Time-Of-Use) race condition and requires 3 trips to the database.
   - **Suggestion:** Since Phase 2 added unique indexes on `email` and `username`, simply attempt the `CreateUser` insert first. Catch the resulting error, check if it's a `pgx.PgError` with code `23505` (unique_violation), and inspect the constraint name to return either `EMAIL_TAKEN` or `USERNAME_TAKEN`.

3. **Medium: Hardcoded Expiry in Token Refresh**
   - **Reference:** `internal/handler/auth.go` (`Refresh`)
   - **Issue:** The handler hardcodes `ExpiresIn: 3600` instead of reading the dynamic TTL from the application config or the service.
   - **Suggestion:** Have `AuthService.RefreshAccessToken` return the `*AuthTokens` struct or the TTL duration alongside the new token.

4. **Medium: Refresh Token Rotation**
   - **Reference:** `internal/service/auth.go` (`RefreshAccessToken`)
   - **Issue:** The current refresh flow issues a new access token but reuses the same refresh token forever (until it expires).
   - **Suggestion:** For better security (Refresh Token Rotation), `RefreshAccessToken` should call `s.GenerateTokens(claims.UserID)` and issue a brand new pair of access and refresh tokens.

---

### 🛠️ Implementation Stages (Low-Level Summary)

For tracking purposes, here is the state of the implementation achieved in this PR:

1. **Stage 1: Core JWT Utility:** Engineered a robust, independent `pkg/jwt` library. Configured it to sign and verify both access and refresh tokens via standard `HMAC-SHA256` signatures, wrapping `golang-jwt/v5`.
2. **Stage 2: Gin Middleware Context:** Implemented the `Auth` middleware to intercept the `Authorization: Bearer` header, validate access token claims, and securely inject the parsed `user_id` into the Gin context for downstream operations.
3. **Stage 3: Secure Service Layer:** Developed `AuthService` to orchestrate secure password hashing via `bcrypt.DefaultCost`, handle registration constraint validation, and broker token issuance flows.
4. **Stage 4: Handlers & Safe DTOs:** Wired up the HTTP layer (`internal/handler/auth.go`, `internal/handler/users.go`). Crucially introduced dedicated request/response Data Transfer Objects (e.g., `UserProfileResponse`, `RegisterResponse`) to guarantee that sensitive internal fields (like `password_hash`) are stripped and never leak into API responses.