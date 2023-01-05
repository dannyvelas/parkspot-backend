CREATE TABLE IF NOT EXISTS admin(
  id TEXT PRIMARY KEY UNIQUE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  is_privileged BOOLEAN NOT NULL,
  token_version INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS resident(
  id CHAR(8) PRIMARY KEY UNIQUE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  phone VARCHAR(20) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  unlim_days BOOLEAN NOT NULL DEFAULT FALSE,
  amt_parking_days_used SMALLINT NOT NULL DEFAULT 0,
  token_version INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS car(
  id UUID PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  license_plate VARCHAR(10) UNIQUE NOT NULL,
  color TEXT NOT NULL,
  make TEXT,
  model TEXT,
  amt_parking_days_used SMALLINT NOT NULL DEFAULT 0
);

-- we are purposely NOT adding a `car`.id foreign key here
-- we don't want changes to a given car to affect the history of permits created
-- thus, we want the car information in each permit to be a "snapshot", at the time the permit was created
CREATE TABLE IF NOT EXISTS permit(
  id SERIAL PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  car_id UUID NOT NULL,
  license_plate VARCHAR(10) NOT NULL,
  color TEXT NOT NULL,
  make TEXT,
  model TEXT,
  start_ts BIGINT NOT NULL,
  end_ts BIGINT NOT NULL,
  request_ts BIGINT,
  affects_days BOOLEAN NOT NULL,
  exception_reason TEXT
);

CREATE TABLE IF NOT EXISTS visitor(
  id UUID PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  relationship TEXT NOT NULL,
  access_start BIGINT NOT NULL,
  access_end BIGINT NOT NULL
);
