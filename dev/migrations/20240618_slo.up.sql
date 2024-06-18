create table if not exists slo (
  id uuid not null primary key,
  name varchar(255) not null unique,
  created_at timestamp not null,
  description varchar(255),
  labels jsonb,
  objective real not null,
);
--;;
create table if not exists slo_records_aggregated (
  name varchar(255) not null,
  started_at timestamp not null,
  success boolean not null,
  value bigint,
  CONSTRAINT fk_slo_name
    FOREIGN KEY(name)
    REFERENCES slo(name)
);
--;;
CREATE INDEX IF NOT EXISTS idx_slo_records_aggregated_name_created_at ON slo_records_aggregated(name, created_at DESC);
--;;
