create table if not exists services (
    id            bigserial primary key,
    version       integer not null default 1,
    created_at    timestamp(8) with time zone not null default now(),
    updated_at    timestamp(8) with time zone not null default now(),
    admin_email   citext,
    password_hash bytea not null,
    name          citext unique not null,
    description   text,
    suspended     bool  not null default false,
    deleted       bool  not null default false
);
