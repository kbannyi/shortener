CREATE TABLE IF NOT EXISTS url
(
    id           varchar(100) PRIMARY KEY,
    short_url    text UNIQUE,
    original_url text
);