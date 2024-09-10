CREATE TABLE
    IF NOT EXISTS users (id serial primary key, email varchar(255));

CREATE TABLE
    IF NOT EXISTS refresh_tokens (payload varchar(255));