-- migrations/0001_init.down.sql

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS data_type;