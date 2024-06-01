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
create table if not exists healthcheck_result (
  id uuid not null primary key,
  success boolean not null,
  labels jsonb,
  created_at timestamp not null,
  summary text not null,
  message text not null,
  healthcheck_id uuid not null,
  foreign key (healthcheck_id) REFERENCES healthcheck(id)
);
--;;
CREATE INDEX IF NOT EXISTS idx_healthcheck_result_created_at ON healthcheck_result(created_at);
--;;
CREATE INDEX IF NOT EXISTS idx_healthcheck_result_healthcheck_id ON healthcheck_result(healthcheck_id);
--;;
