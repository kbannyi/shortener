CREATE TABLE url
(
    id           varchar(100) PRIMARY KEY,
    short_url    text UNIQUE,
    original_url text
);