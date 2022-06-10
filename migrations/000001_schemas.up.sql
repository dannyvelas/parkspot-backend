BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS admin(
  id UUID PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  is_privileged BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS resident(
  id CHAR(8) PRIMARY KEY UNIQUE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  phone VARCHAR(20) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  unlim_days BOOLEAN NOT NULL DEFAULT FALSE,
  amt_parking_days_used SMALLINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS car(
  id UUID PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  license_plate VARCHAR(10) UNIQUE NOT NULL,
  color TEXT NOT NULL,
  make TEXT,
  model TEXT,
  amt_parking_days_used SMALLINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS permit(
  id SERIAL PRIMARY KEY UNIQUE NOT NULL,
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  car_id UUID REFERENCES car(id) ON DELETE CASCADE NOT NULL,
  start_ts BIGINT NOT NULL,
  end_ts BIGINT NOT NULL,
  request_ts BIGINT,
  affects_days BOOLEAN NOT NULL,
  exception_reason TEXT
);



CREATE TYPE relationship AS ENUM('fam/fri', 'contractor');
CREATE TABLE IF NOT EXISTS visitor(
  id UUID PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  resident_id CHAR(8) REFERENCES resident(id) ON DELETE CASCADE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  relationship relationship NOT NULL,
  access_start BIGINT NOT NULL,
  access_end BIGINT NOT NULL
);

COMMIT;
