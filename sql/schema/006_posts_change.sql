-- +goose Up
ALTER TABLE posts 
ALTER COLUMN description TYPE TEXT;



-- +goose Down
ALTER TABLE posts ALTER COLUMN description TYPE VARCHAR(255);