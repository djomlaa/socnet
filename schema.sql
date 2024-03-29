CREATE TABLE IF NOT EXISTS socnet.users (
    id SERIAL NOT NULL PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    username VARCHAR NOT NULL UNIQUE,
    avatar VARCHAR,
    followers_count INT NOT NULL DEFAULT 0 CHECK (followers_count >=0),
    followees_count INT NOT NULL DEFAULT 0 CHECK (followees_count >=0)
);

CREATE TABLE IF NOT EXISTS socnet.follows (
    follower_id INT NOT NULL REFERENCES socnet.users(id),
    followee_id INT NOT NULL REFERENCES socnet.users(id),
    PRIMARY KEY (follower_id, followee_id)
);

CREATE TABLE IF NOT EXISTS socnet.posts (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL  REFERENCES socnet.users(id),
    content VARCHAR NOT NULL,
    spoiler_of VARCHAR,
    nsfw BOOLEAN NOT NULL,
    likes_count INT NOT NULL DEFAULT 0 CHECK (likes_count >=0)
    comments_count INT NOT NULL DEFAULT 0 CHECK (comments_count >=0)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sorted_posts ON socnet.posts (created_at DESC);

CREATE TABLE IF NOT EXISTS socnet.timeline (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES socnet.users(id),
    post_id INT NOT NULL REFERENCES socnet.posts(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS timeline_unique ON socnet.timeline (user_id, post_id);

CREATE TABLE IF NOT EXISTS socnet.post_likes (
    user_id INT NOT NULL REFERENCES socnet.users(id),
    post_id INT NOT NULL REFERENCES socnet.posts(id),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE IF NOT EXISTS socnet.comments (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL  REFERENCES socnet.users(id),
    post_id INT NOT NULL REFERENCES socnet.posts(id),
    content VARCHAR NOT NULL,
    likes_count INT NOT NULL DEFAULT 0 CHECK (likes_count >=0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sorted_comments ON socnet.comments (created_at DESC);

CREATE TABLE IF NOT EXISTS socnet.comment_likes (
    user_id INT NOT NULL REFERENCES socnet.users(id),
    comment_id INT NOT NULL REFERENCES socnet.comments(id),
    PRIMARY KEY (user_id, comments_id)
);

CREATE TABLE IF NOT EXISTS socnet.notifications (
    id SERIAL NOT NULL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES socnet.users(id),
    actors VARCHAR[] NOT NULL,
    type VARCHAR NOT NULL,
    read BOOLEAN NOT NULL DEFAULT false,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sorted_notifications ON socnet.notifications (issued_at DESC);


INSERT INTO socnet.users (id, email, username) VALUES
(1, 'mladen@example.org', 'mladen'),
(2, 'milutin@example.org', 'milutin'),
(3, 'momcilo@example.org', 'momcilo');

INSERT INTO socnet.posts (id, user_id, content, comments_count) VALUES
(1, 1, 'sample post', 1);

INSERT INTO socnet.timeline (id, user_id, post_id) VALUES
(1, 1, 1);

INSERT INTO socnet.comments (id, user_id, post_id, content) VALUES
(1, 1, 1, 'sample post');