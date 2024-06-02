create table if not exists healthcheck (
  id uuid not null primary key,
  random_id integer not null,
  name varchar(255) not null unique,
  description varchar(255),
  interval varchar(255) not null,
  timeout varchar(255) not null,
  labels jsonb,
  enabled boolean not null,
  created_at timestamp not null,
  type varchar(255) not null,
  definition jsonb not null
);
--;;
CREATE INDEX IF NOT EXISTS idx_healthcheck_type ON healthcheck(type);
--;;
