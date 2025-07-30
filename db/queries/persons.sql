-- name: CreatePerson :one
INSERT INTO persons (id, name)
VALUES ($1, $2)
RETURNING id, name, created_at, updated_at;

-- name: GetPersonByID :one
SELECT id, name, created_at, updated_at
FROM persons
WHERE id = $1;

-- name: ListPersons :many
SELECT id, name, created_at, updated_at
FROM persons
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPersons :one
SELECT COUNT(*) FROM persons;

-- name: UpdatePerson :one
UPDATE persons
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, created_at, updated_at;

-- name: DeletePerson :exec
DELETE FROM persons
WHERE id = $1;

-- name: GetPersonByName :one
SELECT id, name, created_at, updated_at
FROM persons
WHERE name = $1;

-- name: SearchPersonsByName :many
SELECT id, name, created_at, updated_at
FROM persons
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name
LIMIT $2 OFFSET $3;
