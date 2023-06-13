create table if not exists users (
    id bigserial primary key,
    created timestamp(8) with time zone not null default now(),
    updated timestamp(8) with time zone not null default now(),
    name text not null,
    friendly_name text,
    email citext unique not null,
    password_hash bytea not null,
    birth_date date,
    gender text,
    nationality text,
    time_zone text,
    activated bool not null default false,
    delegted bool not null default false,
    version integer not null default 1
);
