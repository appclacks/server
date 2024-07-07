create table if not exists pushgateway_metric (
  id uuid not null primary key,
  name varchar(255) not null,
  description varchar(255),
  ttl varchar(255),
  labels jsonb default '{}'::jsonb,
  value double precision,
  type varchar(255),
  created_at timestamp not null,
  expires_at timestamp
);
--;;
CREATE INDEX IF NOT EXISTS idx_pushgateway_expires_at ON pushgateway_metric(expires_at);
--;;
CREATE INDEX IF NOT EXISTS idx_pushgateway_name ON pushgateway_metric(name);
--;;
