CREATE TABLE IF NOT EXISTS socnet.users (
    id SERIAL NOT NULL PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    username VARCHAR NOT NULL UNIQUE,
    followers_count INT NOT NULL DEFAULT 0 CHECK (followers_count >=0),
    followees_count INT NOT NULL DEFAULT 0 CHECK (followees_count >=0)
);

CREATE TABLE IF NOT EXISTS socnet.follows (
    follower_id INT NOT NULL,
    followee_id INT NOT NULL,
    PRIMARY KEY (follower_id, followee_id)
);

INSERT INTO socnet.users (id, email, username) VALUES
(1, 'mladen@example.org', 'mladen'),
(2, 'milutin@example.org', 'milutin'),
(3, 'momcilo@example.org', 'momcilo');