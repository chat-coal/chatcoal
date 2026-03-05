-- +goose Up
ALTER TABLE users ADD COLUMN home_instance VARCHAR(255) DEFAULT NULL;

CREATE TABLE federated_instances (
  id BIGINT PRIMARY KEY,
  domain VARCHAR(255) UNIQUE NOT NULL,
  public_key TEXT NOT NULL,
  name VARCHAR(100),
  created_at DATETIME(3),
  updated_at DATETIME(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE instance_config (
  id INT PRIMARY KEY DEFAULT 1,
  private_key TEXT NOT NULL,
  public_key TEXT NOT NULL,
  domain VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
ALTER TABLE users DROP COLUMN home_instance;
DROP TABLE IF EXISTS federated_instances;
DROP TABLE IF EXISTS instance_config;
