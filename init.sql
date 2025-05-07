CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS agricultural_areas (
  id SERIAL PRIMARY KEY,
  crop_type TEXT NOT NULL,
  planting_date DATE,
  harvest_date DATE,
  geom GEOMETRY(POLYGON, 4326)
);


CREATE TABLE IF NOT EXISTS chat_messages (
  id SERIAL PRIMARY KEY,
  sender_role TEXT NOT NULL,
  content TEXT NOT NULL,
  sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);