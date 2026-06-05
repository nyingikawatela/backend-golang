-- name: CreateUser :one
INSERT INTO USERS (id, created_at, updated_at, email, hashed_password) VALUES (gen_random_uuid(), NOW(), NOW(), $1, $2)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM USERS;

-- name: GetUserById :one
SELECT * FROM USERS WHERE email = $1;