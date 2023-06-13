create table if not exists tokens (
    hash       bytea primary key,
    user_id    bigint not null references users(id)    on delete cascade,
    service_id bigint not null references services(id) on delete cascade,
    expiry     timestamp(0) with time zone not null,
    scope      text   not null
)
