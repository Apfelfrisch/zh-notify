create table events
(
    id INTEGER not null constraint events_pk primary key,
    name TEXT not null,
    place TEXT not null,
    status TEXT not null,
    link TEXT UNIQUE not null,
    date DATETIME not null,
    artist TEXT,
    category TEXT,
    artist_url TEXT,
    artist_img_url TEXT,
    reported_at_new DATETIME,
    reported_at_upcoming DATETIME,
    created_at TIMESTAMP not null DEFAULT CURRENT_TIMESTAMP
);
