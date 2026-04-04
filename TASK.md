# TrackRide Backend MVP Implementation Plan

> Source spec: `/home/broo/Documents/myRiders BE/CLAUDE.md`
> Assumption: backend codebase will live under `/home/broo/Documents/myRiders BE/`

**Goal:** Membangun backend MVP TrackRide dengan auth JWT, PostgreSQL, WebSocket GPS streaming, social feed, leaderboard mingguan, dan notifikasi push.

**Architecture:** Backend dibangun sebagai service Go modular dengan pemisahan `handler`, `service`, `middleware`, `db/sqlc`, `websocket`, dan `jobs`. Fokus implementasi dibagi per domain agar tiap fase bisa diuji dan digabung tanpa menunggu seluruh MVP selesai.

**Tech Stack:** Go 1.22+, Gin, PostgreSQL 16, sqlc, pgx/v5, Redis, gorilla/websocket, robfig/cron, JWT.

---

## Working Notes

- [ ] Gunakan root backend `/home/broo/Documents/myRiders BE/` sebagai basis semua path di bawah
- [ ] Pertahankan format error `{ "error": "ERROR_CODE" }`
- [ ] Selalu tambahkan validasi ownership untuk resource milik user
- [ ] Prioritaskan flow inti MVP lebih dulu: auth, vehicles, rides, GPS, leaderboard
- [ ] Sisihkan keputusan yang masih terbuka ke bagian `Open Decisions`, jangan blok implementasi awal

## Target File Map

### Bootstrap and wiring

- [ ] Create `cmd/server/main.go`
- [ ] Create `internal/config/config.go`
- [ ] Create `internal/router/router.go`
- [ ] Create `internal/middleware/auth.go`
- [ ] Create `internal/middleware/cors.go`

### Database and queries

- [ ] Create `internal/db/migrations/001_init.sql`
- [ ] Create `internal/db/migrations/002_indexes.sql`
- [ ] Create `internal/db/queries/auth.sql`
- [ ] Create `internal/db/queries/users.sql`
- [ ] Create `internal/db/queries/vehicles.sql`
- [ ] Create `internal/db/queries/rides.sql`
- [ ] Create `internal/db/queries/social.sql`
- [ ] Create `internal/db/queries/leaderboard.sql`
- [ ] Create `internal/db/sqlc/` via `sqlc generate`
- [ ] Create `sqlc.yaml`

### Domain handlers and services

- [ ] Create `internal/handler/auth.go`
- [ ] Create `internal/handler/users.go`
- [ ] Create `internal/handler/vehicles.go`
- [ ] Create `internal/handler/rides.go`
- [ ] Create `internal/handler/social.go`
- [ ] Create `internal/handler/leaderboard.go`
- [ ] Create `internal/service/auth.go`
- [ ] Create `internal/service/rides.go`
- [ ] Create `internal/service/metrics.go`
- [ ] Create `internal/service/geocoding.go`
- [ ] Create `internal/service/notifications.go`

### Realtime and jobs

- [ ] Create `internal/websocket/buffer.go`
- [ ] Create `internal/websocket/hub.go`
- [ ] Create `internal/jobs/leaderboard.go`

### Shared packages

- [ ] Create `pkg/jwt/jwt.go`
- [ ] Create `pkg/polyline/polyline.go`

### Tests and docs

- [ ] Create `internal/handler/` tests per domain as endpoints stabilize
- [ ] Create `internal/service/` tests for metrics and ride completion flow
- [ ] Create `pkg/jwt/jwt_test.go`
- [ ] Create `pkg/polyline/polyline_test.go`
- [ ] Update `.env.example` if actual variables change during implementation

---

## Phase 1 - Foundation and Bootstrap

### Story 1.1 Backend project scaffold
Dependencies: none

- [ ] Inisialisasi module Go sesuai namespace final project
- [ ] Tambahkan dependency inti dari spec: Gin, JWT, WebSocket, Redis, pgx, cron, godotenv, uuid, bcrypt
- [ ] Buat struktur folder `cmd/`, `internal/`, `pkg/`
- [ ] Tambahkan placeholder file minimal agar layout backend jelas
- [ ] Rapikan `.env.example` agar sinkron dengan spec backend
- [ ] Pastikan `go mod tidy` berjalan bersih

### Story 1.2 Config loader and environment validation
Dependencies: Story 1.1

- [ ] Buat `internal/config/config.go`
- [ ] Tambahkan `Config` struct untuk semua env yang dibutuhkan
- [ ] Implement `Load()` untuk membaca `.env`
- [ ] Validasi env wajib: database, redis, JWT secrets
- [ ] Tambahkan fallback untuk `PORT`, access TTL, refresh TTL
- [ ] Parse `WS_TOKEN_TTL` dari env agar tidak hardcoded
- [ ] Siapkan helper `mustEnv` dan `getEnv`

### Story 1.3 App bootstrap and graceful startup
Dependencies: Story 1.2

- [ ] Buat `cmd/server/main.go`
- [ ] Inisialisasi config, PostgreSQL pool, Redis client, dan `sqlc Queries`
- [ ] Tambahkan ping awal ke PostgreSQL dan Redis
- [ ] Inisialisasi service, handler, GPS buffer, WebSocket hub, dan leaderboard job
- [ ] Register router dan jalankan server di port config
- [ ] Tambahkan graceful shutdown untuk HTTP server dan koneksi DB/Redis
- [ ] Pastikan log startup memuat environment dasar tanpa membocorkan secret

### Story 1.4 Base middleware and shared response pattern
Dependencies: Story 1.3

- [ ] Buat `internal/middleware/cors.go`
- [ ] Tentukan aturan CORS default untuk mobile/web client
- [ ] Tambahkan helper internal untuk response error konsisten
- [ ] Pastikan `401`, `403`, `404`, `409`, `422`, dan `500` mengikuti error code spec
- [ ] Verifikasi middleware tidak mengubah response WebSocket endpoint

---

## Phase 2 - Database Foundation

### Story 2.1 Initial database schema
Dependencies: Story 1.1

- [ ] Buat `internal/db/migrations/001_init.sql`
- [ ] Tambahkan extension `pgcrypto`
- [ ] Tambahkan enum `vehicle_type`
- [ ] Tambahkan enum `ride_status`
- [ ] Buat tabel `users`
- [ ] Buat tabel `vehicles`
- [ ] Buat tabel `rides`
- [ ] Buat tabel `ride_gps_points`
- [ ] Buat tabel `follows`
- [ ] Buat tabel `ride_likes`
- [ ] Buat tabel `ride_comments`
- [ ] Buat tabel `leaderboard_entries`
- [ ] Review foreign key dan `ON DELETE` behavior

### Story 2.2 Database indexes
Dependencies: Story 2.1

- [ ] Buat `internal/db/migrations/002_indexes.sql`
- [ ] Tambahkan index rides by user untuk ride completed
- [ ] Tambahkan index vehicle lookup
- [ ] Tambahkan index GPS by ride and time
- [ ] Tambahkan index follows by follower
- [ ] Tambahkan index leaderboard by period and rank
- [ ] Tinjau apakah perlu unique/index tambahan untuk query auth dan profile

### Story 2.3 sqlc configuration and query generation
Dependencies: Story 2.1

- [ ] Buat `sqlc.yaml`
- [ ] Set schema ke `internal/db/migrations/`
- [ ] Set query source ke `internal/db/queries/`
- [ ] Set output Go package ke `internal/db/sqlc`
- [ ] Jalankan `sqlc generate`
- [ ] Pastikan generated code masuk path yang konsisten

### Story 2.4 Auth and user queries
Dependencies: Story 2.3

- [ ] Buat `internal/db/queries/auth.sql`
- [ ] Tambahkan query `CreateUser`
- [ ] Tambahkan query `GetUserByEmail`
- [ ] Tambahkan query `GetUserByUsername`
- [ ] Tambahkan query `GetUserByID`
- [ ] Tambahkan query `UpdateUserProfile`
- [ ] Regenerate `sqlc`
- [ ] Verifikasi type hasil query cocok untuk handler auth dan users

### Story 2.5 Vehicle, ride, social, and leaderboard queries
Dependencies: Story 2.3

- [ ] Buat `internal/db/queries/vehicles.sql`
- [ ] Tambahkan query list, create, update, delete vehicle milik user
- [ ] Buat `internal/db/queries/rides.sql`
- [ ] Tambahkan query `CreateRide`
- [ ] Tambahkan query `GetRideByID`
- [ ] Tambahkan query `GetActiveRide`
- [ ] Tambahkan query `UpdateRideCompleted`
- [ ] Tambahkan query `ListRidesByUser`
- [ ] Tambahkan query `InsertGPSPointsBatch`
- [ ] Tambahkan query `GetGPSPointsByRide`
- [ ] Buat `internal/db/queries/social.sql`
- [ ] Tambahkan query follow, unfollow, like, comment, dan feed
- [ ] Buat `internal/db/queries/leaderboard.sql`
- [ ] Tambahkan query compute, delete, insert, dan list leaderboard
- [ ] Regenerate `sqlc`

---

## Phase 3 - Authentication and Users

### Story 3.1 JWT helper package
Dependencies: Story 1.2

- [ ] Buat `pkg/jwt/jwt.go`
- [ ] Definisikan `Claims` dengan `sub` dan `type`
- [ ] Implement `GenerateAccessToken`
- [ ] Implement `GenerateRefreshToken`
- [ ] Implement `ParseToken`
- [ ] Gunakan expiry dari config, bukan angka hardcoded bila wiring config sudah siap
- [ ] Tambahkan test untuk token valid, expired, dan signature invalid

### Story 3.2 Auth middleware
Dependencies: Story 3.1

- [ ] Buat `internal/middleware/auth.go`
- [ ] Baca header `Authorization`
- [ ] Validasi format `Bearer <token>`
- [ ] Parse access token dengan secret yang benar
- [ ] Tolak token invalid atau expired dengan error code sesuai spec
- [ ] Simpan `user_id` ke Gin context
- [ ] Tambahkan helper `GetUserID`

### Story 3.3 Auth service and handler wiring
Dependencies: Story 2.4, Story 3.1

- [ ] Buat `internal/service/auth.go` untuk operasi auth reusable
- [ ] Buat `internal/handler/auth.go`
- [ ] Implement payload validation untuk `register`
- [ ] Cek konflik email saat register
- [ ] Cek konflik username saat register
- [ ] Hash password dengan bcrypt cost yang konsisten
- [ ] Simpan user baru via `CreateUser`
- [ ] Generate access dan refresh token
- [ ] Implement endpoint `login`
- [ ] Implement endpoint `refresh`
- [ ] Implement endpoint `logout`
- [ ] Tentukan apakah refresh token disimpan di Redis untuk invalidation awal atau cukup stateless untuk MVP

### Story 3.4 User profile endpoints
Dependencies: Story 2.4, Story 3.2

- [ ] Buat `internal/handler/users.go`
- [ ] Implement `GET /users/me`
- [ ] Implement `PUT /users/me`
- [ ] Batasi field update hanya untuk profil yang diizinkan
- [ ] Implement `GET /users/:id`
- [ ] Pastikan response tidak membocorkan `password_hash`
- [ ] Tambahkan validasi profile lookup ketika user tidak ditemukan

### Story 3.5 Router auth and user routes
Dependencies: Story 3.3, Story 3.4

- [ ] Buat `internal/router/router.go`
- [ ] Register public auth routes: register, login, refresh
- [ ] Register protected auth route: logout
- [ ] Register protected user routes: me, update me, profile by id
- [ ] Terapkan middleware auth hanya ke grup route yang tepat
- [ ] Pastikan semua handler terhubung ke dependency yang benar

---

## Phase 4 - Vehicle Management

### Story 4.1 Vehicle domain queries and rules
Dependencies: Story 2.5, Story 3.2

- [ ] Review query vehicle agar selalu filter berdasarkan `user_id`
- [ ] Tambahkan validasi enum `vehicle_type`
- [ ] Tentukan rule `is_active` saat user memiliki banyak kendaraan
- [ ] Pastikan delete vehicle menolak atau menangani ride aktif bila dibutuhkan

### Story 4.2 Vehicle handlers
Dependencies: Story 4.1

- [ ] Buat `internal/handler/vehicles.go`
- [ ] Implement `GET /vehicles`
- [ ] Implement `POST /vehicles`
- [ ] Implement `PUT /vehicles/:id`
- [ ] Implement `DELETE /vehicles/:id`
- [ ] Validasi payload `name`, `type`, `brand`, `color`
- [ ] Pastikan update dan delete hanya untuk vehicle milik user login

### Story 4.3 Vehicle route integration and smoke test
Dependencies: Story 4.2, Story 3.5

- [ ] Daftarkan semua route vehicle di router
- [ ] Jalankan smoke test create, list, update, delete manual
- [ ] Verifikasi error `VEHICLE_NOT_FOUND` dipakai secara konsisten

---

## Phase 5 - Ride Lifecycle and Metrics

### Story 5.1 Ride start flow
Dependencies: Story 2.5, Story 3.2, Story 4.2

- [ ] Buat `internal/service/rides.go`
- [ ] Implement `StartRide`
- [ ] Validasi vehicle ada dan dimiliki user
- [ ] Validasi tidak ada ride aktif lain milik user
- [ ] Simpan ride baru dengan status `active`
- [ ] Kembalikan data ride minimum untuk handler

### Story 5.2 Ride handlers for start, list, detail
Dependencies: Story 5.1

- [ ] Buat `internal/handler/rides.go`
- [ ] Implement `POST /rides/start`
- [ ] Generate `ws_token` saat start ride
- [ ] Simpan `ws_token` ke Redis dengan TTL config
- [ ] Implement `GET /rides`
- [ ] Tambahkan pagination `page` dan `limit`
- [ ] Tambahkan filter `vehicle_type` bila query disediakan
- [ ] Implement `GET /rides/:id`
- [ ] Pastikan detail ride hanya bisa diakses owner atau sesuai rule product

### Story 5.3 GPS buffer for batch inserts
Dependencies: Story 2.5

- [ ] Buat `internal/websocket/buffer.go`
- [ ] Definisikan `GPSPoint`
- [ ] Definisikan struct buffer per ride
- [ ] Implement `Add`
- [ ] Implement auto flush saat batch size tercapai
- [ ] Implement timer flush periodik
- [ ] Implement `flush` ke PostgreSQL via `InsertGPSPointsBatch`
- [ ] Implement `FlushAndClear`
- [ ] Pastikan concurrency aman saat add dan flush bersamaan

### Story 5.4 WebSocket hub for GPS streaming
Dependencies: Story 5.2, Story 5.3

- [ ] Buat `internal/websocket/hub.go`
- [ ] Validasi `ws_token` dari Redis
- [ ] Pastikan token cocok dengan `ride_id`
- [ ] Upgrade koneksi ke WebSocket
- [ ] Handle message `ping`
- [ ] Handle message `gps_point`
- [ ] Validasi latitude dan longitude
- [ ] Parse timestamp atau fallback ke UTC now
- [ ] Simpan point ke GPS buffer
- [ ] Kirim `ack` dengan jumlah point yang diterima
- [ ] Flush sisa buffer saat socket disconnect

### Story 5.5 Ride metric computation
Dependencies: Story 2.5, Story 5.3

- [ ] Buat `internal/service/metrics.go`
- [ ] Implement hitung jarak antar titik
- [ ] Implement hitung durasi
- [ ] Implement hitung max speed
- [ ] Implement hitung average speed
- [ ] Implement hitung elevasi total
- [ ] Implement estimasi kalori dasar
- [ ] Tambahkan guard untuk data point kurang dari 2

### Story 5.6 Polyline and route summary
Dependencies: Story 5.5

- [ ] Buat `pkg/polyline/polyline.go`
- [ ] Implement encode polyline
- [ ] Implement decode polyline
- [ ] Tambahkan test encode/decode roundtrip
- [ ] Implement helper downsampling maksimal 500 titik
- [ ] Implement builder `route_summary` untuk polyline, bounding box, dan cities placeholder
- [ ] Tandai dependency reverse geocoding sebagai optional untuk MVP awal

### Story 5.7 Ride stop flow
Dependencies: Story 5.4, Story 5.5, Story 5.6

- [ ] Lengkapi `StopRide` di `internal/service/rides.go`
- [ ] Ambil active ride berdasarkan `ride_id` dan `user_id`
- [ ] Ambil GPS points dari database
- [ ] Hitung metrics final
- [ ] Bangun `route_summary`
- [ ] Simpan hasil ke `UpdateRideCompleted`
- [ ] Ubah status ride menjadi `completed`
- [ ] Implement `POST /rides/:id/stop`
- [ ] Pastikan ride tanpa cukup GPS point tetap bisa ditutup aman

---

## Phase 6 - Social Features

### Story 6.1 Follow and unfollow users
Dependencies: Story 2.5, Story 3.2

- [ ] Buat `internal/handler/social.go`
- [ ] Implement `POST /users/:id/follow`
- [ ] Implement `DELETE /users/:id/follow`
- [ ] Tolak follow diri sendiri
- [ ] Pastikan duplicate follow tidak membuat error tak terduga

### Story 6.2 Ride likes and comments
Dependencies: Story 6.1

- [ ] Implement `POST /rides/:id/like`
- [ ] Tentukan behavior like kedua kali: idempotent atau conflict
- [ ] Implement `POST /rides/:id/comments`
- [ ] Batasi panjang komentar ke 280 karakter
- [ ] Pastikan hanya ride yang valid yang bisa dilike atau dikomentari

### Story 6.3 Feed endpoint
Dependencies: Story 6.1, Story 6.2

- [ ] Tambahkan query feed completed rides dari akun yang di-follow
- [ ] Implement `GET /feed`
- [ ] Tambahkan pagination sederhana bila dibutuhkan
- [ ] Pastikan item feed memuat user, ride ringkas, dan timestamp

### Story 6.4 Push notification integration
Dependencies: Story 6.2

- [ ] Buat `internal/service/notifications.go`
- [ ] Implement request ke Expo Push API
- [ ] Panggil notifikasi saat like atau comment bila target user punya `push_token`
- [ ] Bungkus error notifikasi agar tidak menggagalkan request utama
- [ ] Catat keputusan synchronous vs goroutine di doc implementasi

---

## Phase 7 - Leaderboard

### Story 7.1 Leaderboard computation queries
Dependencies: Story 2.5

- [ ] Lengkapi query `DeleteLeaderboardEntries`
- [ ] Lengkapi query `ComputeWeeklyRankings`
- [ ] Lengkapi query `InsertLeaderboardEntry`
- [ ] Lengkapi query list global leaderboard
- [ ] Lengkapi query list friends leaderboard
- [ ] Pastikan filter `vehicle_type` mendukung mode semua kendaraan dan per kendaraan

### Story 7.2 Weekly leaderboard cron job
Dependencies: Story 7.1

- [ ] Buat `internal/jobs/leaderboard.go`
- [ ] Implement scheduler dengan timezone `Asia/Jakarta`
- [ ] Jalankan job setiap Senin 00:01 WIB
- [ ] Hapus leaderboard lama untuk periode yang sama sebelum insert ulang
- [ ] Hitung leaderboard untuk semua kendaraan
- [ ] Hitung leaderboard per `motor`
- [ ] Hitung leaderboard per `mobil`
- [ ] Hitung leaderboard per `sepeda`
- [ ] Tambahkan log start, progress, error, dan selesai

### Story 7.3 Leaderboard handlers and routes
Dependencies: Story 7.2, Story 3.2

- [ ] Buat `internal/handler/leaderboard.go`
- [ ] Implement `GET /leaderboard`
- [ ] Implement `GET /leaderboard/friends`
- [ ] Tambahkan dukungan query `period_type` dan `vehicle_type` bila dipakai di product
- [ ] Daftarkan route leaderboard di router

---

## Phase 8 - Hardening, Testing, and Delivery

### Story 8.1 Standardize error handling
Dependencies: Phase 3, Phase 4, Phase 5, Phase 6, Phase 7

- [ ] Audit semua handler agar error code sesuai dokumen spec
- [ ] Ganti raw database error dengan error domain yang aman
- [ ] Pastikan `VALIDATION_ERROR` dipakai untuk bind/validation failure
- [ ] Pastikan `NOT_FOUND` atau error domain spesifik dipakai konsisten
- [ ] Tinjau response body agar tidak membocorkan detail internal

### Story 8.2 Automated tests
Dependencies: Semua phase inti selesai minimal satu alur end-to-end

- [ ] Tambahkan unit test untuk `pkg/jwt`
- [ ] Tambahkan unit test untuk `pkg/polyline`
- [ ] Tambahkan unit test untuk `internal/service/metrics.go`
- [ ] Tambahkan service test untuk `StopRide` dengan sample GPS points
- [ ] Tambahkan handler test untuk auth register dan login
- [ ] Tambahkan handler test untuk vehicle CRUD minimal happy path
- [ ] Tambahkan handler test untuk ride start dan stop
- [ ] Tambahkan integration smoke test untuk WebSocket GPS flow bila memungkinkan

### Story 8.3 Local runbook and developer onboarding
Dependencies: Story 8.2

- [ ] Tambahkan langkah setup lokal PostgreSQL 16 dan Redis
- [ ] Dokumentasikan env wajib
- [ ] Dokumentasikan urutan run migration, `sqlc generate`, dan start server
- [ ] Dokumentasikan endpoint inti untuk QA manual
- [ ] Tambahkan contoh payload auth, vehicle, ride start, gps stream, dan ride stop

### Story 8.4 Release readiness review
Dependencies: Story 8.3

- [ ] Verifikasi semua route pada spec sudah tersedia
- [ ] Verifikasi cron leaderboard berjalan
- [ ] Verifikasi WebSocket auth token expiry berjalan
- [ ] Verifikasi ride completion menghasilkan metrics dan route summary
- [ ] Verifikasi notifikasi tidak memblok request utama
- [ ] Catat scope yang ditunda dari open questions

---

## Suggested Execution Order

- [ ] Mulai dari Phase 1 hingga router dasar hidup
- [ ] Selesaikan Phase 2 sebelum handler domain banyak ditulis
- [ ] Tuntaskan Phase 3 dan Phase 4 agar auth dan vehicle ownership stabil
- [ ] Kerjakan Phase 5 sebagai jalur kritis MVP
- [ ] Tambahkan Phase 6 dan Phase 7 setelah flow ride selesai
- [ ] Tutup dengan Phase 8 untuk quality gate dan handoff

## Open Decisions

- [ ] Putuskan provider reverse geocoding: Google Maps API atau Mapbox
- [ ] Putuskan apakah Redis wajib sejak MVP awal atau hanya untuk ws token
- [ ] Putuskan notifikasi dijalankan synchronous atau background goroutine
- [ ] Putuskan mode deployment: single binary atau Docker Compose
- [ ] Putuskan retention policy untuk `ride_gps_points`

## MVP Done Criteria

- [ ] User bisa register, login, refresh token, dan logout
- [ ] User bisa kelola vehicle sendiri
- [ ] User bisa start ride dan menerima `ws_token`
- [ ] Client bisa stream GPS via WebSocket dan point tersimpan batch ke database
- [ ] User bisa stop ride dan mendapatkan metrics final
- [ ] User bisa melihat history ride dan detail ride
- [ ] User bisa follow, like, comment, dan melihat feed
- [ ] Leaderboard mingguan global dan friends bisa diakses
- [ ] Error format konsisten
- [ ] Setup lokal terdokumentasi dan bisa dijalankan developer lain
