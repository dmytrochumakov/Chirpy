-- +goose Up
ALTER TABLE users
ADD COLUMN is_chirpy_red BOOL NOT NULL DEFAULT false; 

-- +goose Down
ALTER TABLE users
DROP is_chirpy_red;