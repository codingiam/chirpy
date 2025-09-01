-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
           gen_random_uuid(),
           NOW(),
           NOW(),
           $1,
           $2

       )
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
WHERE (sqlc.narg(user_id)::uuid IS NULL OR user_id = sqlc.narg(user_id))
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'ASC' THEN created_at END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'DESC' THEN created_at END DESC;

-- name: GetChirpByID :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps WHERE id = $1;
