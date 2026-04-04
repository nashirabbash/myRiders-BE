-- Users indexes
CREATE UNIQUE INDEX idx_users_email ON users(LOWER(email));
CREATE UNIQUE INDEX idx_users_username ON users(LOWER(username));

-- Vehicles indexes
CREATE INDEX idx_vehicles_user_id ON vehicles(user_id) WHERE is_active = TRUE;

-- Rides indexes
CREATE INDEX idx_rides_user_id ON rides(user_id);
CREATE INDEX idx_rides_vehicle_id ON rides(vehicle_id);
CREATE INDEX idx_rides_user_completed ON rides(user_id, started_at DESC) WHERE status = 'completed';
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_started_at ON rides(started_at DESC);

-- GPS points indexes
CREATE INDEX idx_gps_points_ride_id ON ride_gps_points(ride_id);
CREATE INDEX idx_gps_points_ride_time ON ride_gps_points(ride_id, recorded_at ASC);
CREATE INDEX idx_gps_points_recorded_at ON ride_gps_points(recorded_at DESC);

-- Follows indexes
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);

-- Ride likes indexes
CREATE INDEX idx_ride_likes_ride_id ON ride_likes(ride_id);
CREATE INDEX idx_ride_likes_user_id ON ride_likes(user_id);

-- Ride comments indexes
CREATE INDEX idx_ride_comments_ride_id ON ride_comments(ride_id);
CREATE INDEX idx_ride_comments_user_id ON ride_comments(user_id);
CREATE INDEX idx_ride_comments_created_at ON ride_comments(created_at DESC);

-- Leaderboard indexes
CREATE INDEX idx_leaderboard_user_id ON leaderboard_entries(user_id);
CREATE INDEX idx_leaderboard_period ON leaderboard_entries(period_type, period_start, rank ASC);
CREATE INDEX idx_leaderboard_vehicle_type ON leaderboard_entries(vehicle_type, period_type, period_start);
