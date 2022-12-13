BEGIN;

drop table if exists likes;

drop type if exists like_type;
drop type if exists like_mode;

drop table if exists answers;

drop table if exists questions;

drop index if exists users_email_idx;

drop index if exists users_email_uniq_idx;

drop index if exists sessions_user_idx;

drop table if exists sessions;

drop table if exists users;

drop type if exists user_status;

END;
