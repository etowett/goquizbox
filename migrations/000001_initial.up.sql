BEGIN;

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
  created_at timestamptz  not null default clock_timestamp(),
  updated_at timestamptz
);

create index sessions_user_idx ON sessions(user_id);

create table questions (
  id bigserial primary key,
  user_id bigint references users(id),
  title varchar(255) not null,
  body text,
  tags varchar(255) not null,
  created_at timestamptz  not null default clock_timestamp(),
  updated_at timestamptz
);

create table answers (
  id bigserial primary key,
  user_id bigint references users(id),
  question_id bigint references questions(id),
  body text,
  created_at timestamptz  not null default clock_timestamp(),
  updated_at timestamptz
);

create type like_type as enum ('question', 'answer');
create type like_mode as enum ('up', 'down');

create table likes (
  id bigserial primary key,
  user_id bigint references users(id),
  type_id bigint references questions(id),
  type like_type not null,
  mode like_mode not null,
  created_at timestamptz  not null default clock_timestamp(),
  updated_at timestamptz
);

END;
