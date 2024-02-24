CREATE TABLE IF NOT EXISTS artists (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS albums (
    id INTEGER PRIMARY KEY,
    artist_id INTEGER,
    name TEXT NOT NULL,
    FOREIGN KEY(artist_id) REFERENCES artists(id),
    UNIQUE(artist_id, name)
);

CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY,
    album_id INTEGER,
    title TEXT NOT NULL,
    FOREIGN KEY(album_id) REFERENCES albums(id),
    UNIQUE(album_id, title)
);


-- Still figuring this out.