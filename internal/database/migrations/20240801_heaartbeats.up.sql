create table if not exists heartbeat (
  id uuid not null primary key,
  name varchar(255) not null unique,
  description varchar(255),
  ttl varchar(255),
  labels jsonb default '{}'::jsonb,
  created_at timestamp not null,
  refreshed_at timestamp
);
--;;
CREATE INDEX IF NOT EXISTS idx_heartbeat_refreshed_at ON heartbeat(refreshed_at);
--;;
