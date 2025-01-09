-- +goose Up
CREATE TABLE chatIds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    chatId TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE chatIds;
