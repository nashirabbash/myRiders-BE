## PR Review: Follow-up on Latest Fixes

Thanks for the updates! The fixes for `UpdateMe` (fetching the existing user to preserve fields) and the dynamic TTL for `RefreshResponse` look perfect. Removing the `NotBefore` claim and using the standard `Subject` field for the User ID in JWT are also excellent enhancements!

However, it looks like two of the points from the previous review were slightly misunderstood or only partially addressed. Please take a look at these remaining issues before we merge:

### 🔍 Remaining Findings & Optimizations

1. **High: Race Condition & Extra DB Trips in `Register` (Still Present)**
   - **Reference:** `internal/service/auth.go` (`Register`)
   - **Issue:** The latest commit added `errors.Is(err, pgx.ErrNoRows)` checks, which is good for error handling, but it **did not** remove the two `SELECT` queries before the `INSERT`. This means the TOCTOU (Time-Of-Check to Time-Of-Use) race condition still exists, and we are still making 3 roundtrips to the database instead of 1.
   - **Suggestion:** Delete the `GetUserByEmail` and `GetUserByUsername` checks entirely from the `Register` function. Instead, immediately call `s.queries.CreateUser`. If it fails, check the PostgreSQL error code for a unique violation (`23505`).
   ```go
   user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{...})
   if err != nil {
       var pgErr *pgconn.PgError
       if errors.As(err, &pgErr) && pgErr.Code == "23505" {
           if pgErr.ConstraintName == "idx_users_email" || pgErr.ConstraintName == "users_email_key" {
               return nil, nil, fmt.Errorf("EMAIL_TAKEN")
           }
           if pgErr.ConstraintName == "idx_users_username" || pgErr.ConstraintName == "users_username_key" {
               return nil, nil, fmt.Errorf("USERNAME_TAKEN")
           }
       }
       return nil, nil, err
   }
   ```

2. **Medium: Refresh Token Rotation (Not Implemented)**
   - **Reference:** `internal/service/auth.go` (`RefreshAccessToken`)
   - **Issue:** The `RefreshAccessToken` function was updated to return the dynamic expiry, but it still **does not** issue a new refresh token. It only issues a new access token, meaning the same refresh token is reused indefinitely until it expires.
   - **Suggestion:** Instead of just calling `GenerateAccessToken`, call `s.GenerateTokens(claims.UserID())` inside `RefreshAccessToken` and update the function signature (and the handler) to return both a new access token AND a new refresh token. This rotates the refresh token on every use, which is a critical security best practice.

---

### 🛠️ Implementation Stages (Low-Level Summary)

For tracking:
1. **Stage 1 (Completed):** Built robust JWT generation/parsing utilities utilizing standard claims (`Subject`) and eliminating clock skew issues.
2. **Stage 2 (Completed):** Integrated Gin middleware for protected routes with proper 401 unauthenticated aborts.
3. **Stage 3 (Completed):** Correctly mapped Safe DTOs to strip sensitive fields from API responses and handled partial updates in `UpdateMe` without erasing existing data.
4. **Stage 4 (Pending):** Refactoring `Register` to rely exclusively on database unique constraints (optimizing to 1 DB trip) and implementing proper Refresh Token Rotation in the `/refresh` endpoint.