-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create type user_status as enum ('unverified', 'active', 'inactive');

create table users (
  id bigserial primary key,
  first_name varchar(225) not null,
  last_name varchar(225) not null,
  email varchar(255) not null,
  email_activation_key varchar(10),
  email_verified boolean default false,
  status user_status not null default 'unverified',
  password_hash varchar(255) not null,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz
);

create unique index users_email_uniq_idx ON users(LOWER(email));

create index users_email_idx ON users(LOWER(email));

create table sessions (
  id bigserial primary key,
  user_id bigint references users(id),
  deactivated_at timestamptz null,
  expires_at timestamptz not null,
  ip_address varchar(255) not null,
  last_refreshed_at timestamptz  not null,
  user_agent varchar(255) not null,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz
);

create index sessions_user_idx ON sessions(user_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop index if exists sessions_user_idx;

drop table if exists sessions;

drop index if exists users_email_idx;

drop index if exists users_email_uniq_idx;

drop table if exists users;

drop type if exists user_status;
