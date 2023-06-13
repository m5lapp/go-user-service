create table if not exists permissions (
    id         bigserial primary key,
    service_id bigint not null references services(id) on delete cascade,
    permission text   not null
);

create table if not exists user_permissions (
    user_id       bigint not null references users(id) on delete cascade,
    permission_id bigint not null references permissions(id) on delete cascade,
    primary key (user_id, permission_id)
);

/*insert into permissions
    (service_id, permission)
values
    (1, 'users:write'),
    (1, 'permissions:write'),
    (1, 'tokens:authenticate'),
    (1, 'tokens:create');*/
