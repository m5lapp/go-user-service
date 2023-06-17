create table if not exists users (
    id            bigserial primary key,
    version       integer not null default 1,
    created_at    timestamp(8) with time zone not null default now(),
    updated_at    timestamp(8) with time zone not null default now(),
    email         citext unique not null,
    password_hash bytea not null,
    name          text  not null,
    friendly_name text,
    birth_date    date,
    gender        text,
    country_code  text,
    time_zone     text,
    activated     bool not null default false,
    suspended     bool not null default false,
    deleted       bool not null default false
);
