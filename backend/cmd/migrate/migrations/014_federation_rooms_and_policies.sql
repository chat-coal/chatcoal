-- +goose Up

-- Instance policy allow/block list
CREATE TABLE instance_policies (
  id BIGINT PRIMARY KEY,
  domain VARCHAR(255) NOT NULL UNIQUE,
  policy VARCHAR(10) NOT NULL,
  note VARCHAR(500) DEFAULT '',
  created_by BIGINT NOT NULL,
  created_at DATETIME(3),
  updated_at DATETIME(3)
);

-- Default policy on instance_config
ALTER TABLE instance_config ADD COLUMN default_policy VARCHAR(10) NOT NULL DEFAULT 'open';

-- Federation ID on channels
ALTER TABLE channels ADD COLUMN federation_id VARCHAR(64) DEFAULT NULL UNIQUE;

-- Links: local channel <-> remote federated channel
CREATE TABLE federated_channel_links (
  id BIGINT PRIMARY KEY,
  channel_id BIGINT NOT NULL,
  remote_domain VARCHAR(255) NOT NULL,
  remote_federation_id VARCHAR(64) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  UNIQUE KEY uq_channel_remote (channel_id, remote_domain, remote_federation_id),
  KEY idx_remote_fed (remote_federation_id)
);

-- +goose Down

DROP TABLE IF EXISTS federated_channel_links;
ALTER TABLE channels DROP COLUMN federation_id;
ALTER TABLE instance_config DROP COLUMN default_policy;
DROP TABLE IF EXISTS instance_policies;
