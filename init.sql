CREATE TABLE IF NOT EXISTS "public"."directors" (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS "public"."movies" (
    id SERIAL PRIMARY KEY,
    isbn VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    director_id INTEGER REFERENCES directors(id)
);

INSERT INTO "public"."directors" (first_name, last_name) VALUES
('Uladzislau', 'Kleshchanka'),
('John', 'Doe');

INSERT INTO "public"."movies" (isbn, title, director_id) VALUES
('438227', 'Movie One', 1),
('454551', 'Movie Two', 2);