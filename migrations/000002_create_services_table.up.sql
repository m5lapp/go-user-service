create table if not exists services (
    id            bigserial primary key,
    created       timestamp(8) with time zone not null default now(),
    updated       timestamp(8) with time zone not null default now(),
    version       integer not null default 1,
    name          text not null,
    admin_email   citext,
    password_hash bytea not null,
    suspended     bool not null default false,
    deleted       bool not null default false
);
