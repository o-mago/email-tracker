begin;

create table if not exists users
(
    id         varchar primary key,
    name       varchar     not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
