create table if not exists tokens (
    hash       bytea primary key,
    user_id    bigint references users(id)    on delete cascade,
    service_id bigint references services(id) on delete cascade,
    expiry     timestamp(0) with time zone not null,
    scope      text   not null
    -- Ensure that either user_id is set, or service_id is set, but not both.
    constraint single_owner check ((user_id is null) != (service_id is null))
)
