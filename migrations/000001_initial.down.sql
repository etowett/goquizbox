BEGIN;

drop table if exists answers;

drop table if exists questions;

drop index if exists users_username_idx;
drop index if exists users_email_idx;

drop index if exists users_username_uniq_idx;
drop index if exists users_email_uniq_idx;

drop table if exists users;

drop type if exists user_status;

drop index if exists sessions_user_idx;

drop table if exists sessions;

END;
