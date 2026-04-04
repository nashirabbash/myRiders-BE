# Phase 3 - Authentication and Users

## Objective
Implement a secure, robust authentication system using JWTs and create user profile management endpoints for the TrackRide backend MVP. This phase provides the security layer and user identity management required for all subsequent domain features.

## Task Breakdown

### Story 3.1: JWT Helper Package
- **File to create:** `pkg/jwt/jwt.go`
- **Actions:**
  - Define a custom `Claims` struct that includes standard claims (like `sub` for user ID) and a custom `type` claim (e.g., "access" vs. "refresh").
  - Implement `GenerateAccessToken` using the configured access token TTL and secret.
  - Implement `GenerateRefreshToken` using the configured refresh token TTL and secret.
  - Implement `ParseToken` to validate signatures and expiration.
  - Write unit tests (`pkg/jwt/jwt_test.go`) covering valid tokens, expired tokens, and signature mismatch scenarios.

### Story 3.2: Authentication Middleware
- **File to create:** `internal/middleware/auth.go`
- **Actions:**
  - Create a Gin middleware that extracts the `Authorization` header.
  - Validate the `Bearer <token>` format.
  - Parse and validate the access token using the `pkg/jwt` helper.
  - Return standardized JSON errors for invalid or expired tokens based on the project spec.
  - On success, extract the `user_id` from the token claims and store it in the Gin context.
  - Provide a safe helper function `GetUserID(c *gin.Context) (string, bool)` to retrieve the user ID in subsequent handlers.

### Story 3.3: Auth Service and Handler Wiring
- **Files to create:** `internal/service/auth.go` and `internal/handler/auth.go`
- **Actions:**
  - **Service:** Implement reusable business logic for authentication. Hash passwords using `bcrypt` with a standard cost. 
  - **Handler:** Implement request payload validation for the `register` endpoint.
  - Check the database for email and username conflicts before creating a user.
  - Insert the new user via the `CreateUser` sqlc query.
  - Generate both an access and a refresh token upon successful registration or login.
  - Implement the `login` endpoint (verify password hash vs database hash).
  - Implement the `refresh` endpoint (validate refresh token and issue a new access token).
  - Implement the `logout` endpoint (for MVP, decide if stateless logout is sufficient or if the refresh token should be blacklisted in Redis).

### Story 3.4: User Profile Endpoints
- **File to create:** `internal/handler/users.go`
- **Actions:**
  - Implement `GET /users/me` to fetch the logged-in user's profile using the context's `user_id`.
  - Implement `PUT /users/me` to update the user's profile. Restrict updates to permitted fields only (e.g., `display_name`, `avatar_url`).
  - Implement `GET /users/:id` to fetch a public profile of any user.
  - Ensure all database responses are mapped to a secure DTO that **never** leaks the `password_hash`.
  - Handle cases where a user is not found gracefully with a standardized 404 response.

### Story 3.5: Router Auth and User Routes
- **File to update:** `internal/router/router.go`
- **Actions:**
  - Register public unauthenticated routes: `POST /register`, `POST /login`, `POST /refresh`.
  - Create a protected route group using the Auth Middleware.
  - Register protected routes: `POST /logout`, `GET /users/me`, `PUT /users/me`, `GET /users/:id`.
  - Ensure all handlers are properly wired with their respective service or database dependencies.

---

## Implementation Stages (Low-Level Guide)

1. **Stage 1: Core JWT Utility**
   Start purely in the `pkg/jwt` package. Build this without any Gin dependencies. Use `golang-jwt/jwt/v5`. Keep the interface clean. This is the cryptographic core of the phase. Write tests here immediately to ensure your token generation and parsing are rock solid before integrating with the web layer.

2. **Stage 2: Gin Middleware & Context Context**
   Move to `internal/middleware/auth.go`. The goal here is to bridge the Gin HTTP request with your `pkg/jwt` logic. Be extremely careful with type assertions when retrieving the `user_id` from the Gin context. Always return early via `c.AbortWithStatusJSON()` if authentication fails, preventing downstream handlers from executing.

3. **Stage 3: Service Layer & Password Hashing**
   In `internal/service/auth.go`, isolate the `golang.org/x/crypto/bcrypt` usage. Your handler should not hash passwords directly; it should call the service. Keep database lookups (using the `sqlc` queries generated in Phase 2) inside the service or handler, but separate the HTTP payload binding from the core logic. 

4. **Stage 4: Handlers & Data Transfer Objects (DTOs)**
   Build out the handlers (`internal/handler/auth.go` and `internal/handler/users.go`). Use `gin`'s `ShouldBindJSON` for incoming payloads and apply validation tags. Crucially, define localized struct types (DTOs) for your JSON responses so that the generated `sqlc` structs (which contain the password hash) are never serialized directly to the client.

5. **Stage 5: Router Wiring**
   Finally, wire everything together in `internal/router/router.go`. Apply the middleware exclusively to the endpoints that require identity. Run local curl or Postman tests to ensure the protected endpoints properly reject requests lacking the `Authorization` header.