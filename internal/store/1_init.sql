CREATE TABLE
    IF NOT EXISTS users (
        id serial primary key,
        email varchar(255),
        uuid varchar(255),
        current_ip_sign_in varchar(255),
        last_sign_in_at TIMESTAMPTZ
    );

CREATE TABLE
    IF NOT EXISTS refresh_tokens (payload varchar(255));