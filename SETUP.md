# TrackRide Backend - Local Setup Guide

This guide will help you set up and run the TrackRide backend locally for development and testing.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.22+** - [Download here](https://golang.org/dl/)
- **PostgreSQL 16** - [Download here](https://www.postgresql.org/download/)
- **Redis** - [Download here](https://redis.io/download) or install via package manager
- **Git** - For version control

### Verify Installations

```bash
go version          # Should show 1.22 or higher
psql --version      # Should show 16.x
redis-cli --version # Should show 7.x or higher
```

## Step 1: Clone the Repository

```bash
git clone https://github.com/nashirabbash/myRiders-BE.git
cd myRiders-BE
```

## Step 2: Set Up PostgreSQL

### Start PostgreSQL Service

**Linux/macOS:**
```bash
# If installed via Homebrew
brew services start postgresql

# Verify it's running
psql --version && pg_isready
```

**Windows:**
- Start PostgreSQL from the Start Menu or Services

### Create Database and User

```bash
# Connect to PostgreSQL
psql -U postgres

# In the psql prompt:
CREATE USER trackride WITH PASSWORD 'trackride_dev_password';
CREATE DATABASE trackride OWNER trackride;
GRANT ALL PRIVILEGES ON DATABASE trackride TO trackride;
\q
```

## Step 3: Set Up Redis

### Start Redis Service

**Linux/macOS:**
```bash
# If installed via Homebrew
brew services start redis

# Verify it's running
redis-cli ping  # Should respond with PONG
```

**Windows:**
- Start Redis from the Start Menu or Services

**Docker (Alternative):**
```bash
docker run -d -p 6379:6379 redis:latest
```

## Step 4: Configure Environment Variables

Create a `.env` file in the project root:

```bash
cat > .env << 'ENVFILE'
# Database
DATABASE_URL=postgres://trackride:trackride_dev_password@localhost:5432/trackride?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT Secrets (use strong random values in production)
JWT_ACCESS_SECRET=dev_access_secret_min_32_chars_long_string_required
JWT_REFRESH_SECRET=dev_refresh_secret_min_32_chars_long_string_required
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=720h

# WebSocket
WS_TOKEN_TTL=600

# Google Maps API (optional for development)
GOOGLE_MAPS_API_KEY=your_api_key_here

# Server Configuration
PORT=8080
APP_ENV=development
ENVFILE
```

## Step 5: Run Database Migrations

### Install sqlc (if not already installed)

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Generate sqlc Code

```bash
sqlc generate
```

### Run Migrations

```bash
# Connect to the database
psql -U trackride -d trackride -h localhost

# Copy and paste the SQL from internal/db/migrations/001_init.sql
\i internal/db/migrations/001_init.sql
\i internal/db/migrations/002_indexes.sql

# Verify tables were created
\dt
\q
```

## Step 6: Install Dependencies

```bash
go mod download
go mod tidy
```

## Step 7: Build and Run the Server

### Build

```bash
go build -o trackride ./cmd/server
```

### Run

```bash
./trackride
# Or directly
go run ./cmd/server/main.go
```

You should see output like:
```
2026/04/05 08:00:00 [GIN-debug] Loaded HTML Templates (0): 
2026/04/05 08:00:00 [GIN-debug] Listening and serving HTTP on :8080
```

The server is now running at `http://localhost:8080`

## Step 8: Verify the Setup

Test the health endpoint:

```bash
curl http://localhost:8080/v1/health
# Expected response: {"status":"healthy","app":"trackride-backend"}
```

## Testing Core Endpoints

Here are the key endpoints for manual QA testing:

### 1. User Registration

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPassword123",
    "display_name": "Test User"
  }'
```

**Expected Response (201):**
```json
{
  "id": "uuid-string",
  "username": "testuser",
  "email": "test@example.com",
  "display_name": "Test User",
  "avatar_url": "",
  "access_token": "jwt-token",
  "refresh_token": "jwt-token",
  "expires_in": 3600
}
```

Save the `access_token` for further requests.

### 2. User Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123"
  }'
```

### 3. Create a Vehicle

Replace `{ACCESS_TOKEN}` with the token from registration:

```bash
curl -X POST http://localhost:8080/v1/vehicles \
  -H "Authorization: Bearer {ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "sepeda",
    "name": "My Mountain Bike",
    "brand": "Trek",
    "color": "Blue"
  }'
```

**Expected Response (201):**
```json
{
  "id": "uuid-string",
  "type": "sepeda",
  "name": "My Mountain Bike",
  "brand": "Trek",
  "color": "Blue",
  "is_active": true,
  "created_at": "2026-04-05T08:00:00Z",
  "updated_at": "2026-04-05T08:00:00Z"
}
```

Save the vehicle `id` for the next step.

### 4. Start a Ride

```bash
curl -X POST http://localhost:8080/v1/rides/start \
  -H "Authorization: Bearer {ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_id": "{VEHICLE_ID}",
    "title": "Morning Ride"
  }'
```

**Expected Response (201):**
```json
{
  "ride_id": "uuid-string",
  "ws_token": "uuid-string",
  "started_at": "2026-04-05T08:00:00Z"
}
```

Save the `ride_id` and `ws_token`.

### 5. Send GPS Points via WebSocket

Install a WebSocket client tool (e.g., `websocat`):

```bash
# Install websocat
cargo install websocat  # or download binary

# Connect to WebSocket
websocat ws://localhost:8080/v1/rides/{RIDE_ID}/stream?token={WS_TOKEN}

# Send GPS points (JSON format, one per line)
{"type":"gps_point","lat":37.5,"lng":-122.5,"speed_kmh":15.0,"elevation_m":0,"timestamp":"2026-04-05T08:00:00Z"}
{"type":"gps_point","lat":37.501,"lng":-122.501,"speed_kmh":16.0,"elevation_m":10,"timestamp":"2026-04-05T08:01:00Z"}
{"type":"ping"}  # Send a ping to test connection
```

### 6. Stop a Ride

```bash
curl -X POST http://localhost:8080/v1/rides/{RIDE_ID}/stop \
  -H "Authorization: Bearer {ACCESS_TOKEN}" \
  -H "Content-Type: application/json"
```

**Expected Response (200):**
```json
{
  "id": "uuid-string",
  "user_id": "uuid-string",
  "vehicle_id": "uuid-string",
  "started_at": "2026-04-05T08:00:00Z",
  "ended_at": "2026-04-05T08:05:00Z",
  "distance_km": 0.5,
  "duration_seconds": 300,
  "max_speed_kmh": 16.0,
  "avg_speed_kmh": 15.5,
  "elevation_m": 10.0,
  "calories": 45,
  "status": "completed"
}
```

### 7. List User Rides

```bash
curl -X GET "http://localhost:8080/v1/rides?page=1&limit=20" \
  -H "Authorization: Bearer {ACCESS_TOKEN}"
```

### 8. Get Leaderboard

```bash
curl -X GET "http://localhost:8080/v1/leaderboard" \
  -H "Authorization: Bearer {ACCESS_TOKEN}"
```

## Troubleshooting

### PostgreSQL Connection Error

**Error:** `connection refused`

**Solution:**
```bash
# Check if PostgreSQL is running
psql -U postgres -c "SELECT 1"

# If not running, start it
brew services start postgresql  # macOS
sudo systemctl start postgresql  # Linux
```

### Redis Connection Error

**Error:** `connection refused`

**Solution:**
```bash
# Check if Redis is running
redis-cli ping

# If not running, start it
brew services start redis      # macOS
sudo systemctl start redis     # Linux
```

### Port Already in Use

**Error:** `listen tcp :8080: bind: address already in use`

**Solution:**
```bash
# Find and kill the process using port 8080
lsof -i :8080
kill -9 <PID>

# Or change the PORT in .env
PORT=8081
```

## Development Tips

- **Code Generation:** Run `sqlc generate` after modifying SQL queries
- **Database Reset:** Drop and recreate the database to start fresh:
  ```bash
  psql -U postgres -c "DROP DATABASE trackride;"
  psql -U postgres -c "CREATE DATABASE trackride OWNER trackride;"
  ```
- **View Logs:** Server logs are printed to stdout
- **Database Queries:** Connect directly with `psql -U trackride -d trackride`
- **Testing:** Run `go test ./...` to execute all tests

## Next Steps

- Read [CLAUDE.md](./CLAUDE.md) for detailed API specifications
- Check [internal/router/router.go](./internal/router/router.go) for all available endpoints
- Review test files for integration examples

## Support

For issues or questions, please:
1. Check existing GitHub issues
2. Review error messages in server logs
3. Consult the inline documentation in relevant Go files
