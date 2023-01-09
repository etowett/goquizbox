-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table questions (
  id bigserial primary key,
  user_id bigint references users(id),
  title varchar(255) not null,
  body text,
  tags varchar(255) not null,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz
);

create table answers (
  id bigserial primary key,
  user_id bigint references users(id),
  question_id bigint references questions(id),
  body text,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz
);

create type vote_type as enum ('question', 'answer');
create type vote_mode as enum ('up', 'down');

create table votes (
  id bigserial primary key,
  user_id bigint references users(id),
  kind_id bigint,
  kind vote_type not null,
  mode vote_mode not null,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists votes;

drop type if exists vote_mode;

drop type if exists vote_type;

drop table if exists answers;

drop table if exists questions;
