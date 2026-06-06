-- name: CreateChirp :one
INSERT INTO chirps (body, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps;

-- name: GetChirpById :one
SELECT * FROM chirps WHERE id = $1;



-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: UpdateRedChirp :one
UPDATE users SET is_chirpy_red = true WHERE id = $1

RETURNING *;