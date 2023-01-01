CREATE TYPE mkind AS ENUM ('counter', 'gauge');

CREATE TABLE IF NOT EXISTS metrics(
    id    varchar(255) primary key,
    name  varchar(255) not null,
    kind  mkind not null,
    value double precision
);

CREATE INDEX metrics__idx ON metrics (id);
