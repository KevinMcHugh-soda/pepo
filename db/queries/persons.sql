-- name: CreatePerson :one
INSERT INTO person (id, name)
VALUES (x2b($1), $2)
RETURNING b2x(id) as id, name, created_at, updated_at;

-- name: GetPersonByID :one
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE id = x2b($1);

-- name: ListPersons :many
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPersons :one
SELECT COUNT(*) FROM person;

-- name: UpdatePerson :one
UPDATE person
SET name = $2, updated_at = NOW()
WHERE id = x2b($1)
RETURNING b2x(id) as id, name, created_at, updated_at;

-- name: DeletePerson :exec
DELETE FROM person
WHERE id = x2b($1);

-- name: GetPersonByName :one
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE name = $1;

-- name: SearchPersonsByName :many
SELECT b2x(id) as id, name, created_at, updated_at
FROM person
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: ListPersonsWithLastAction :many
SELECT
    b2x(p.id) as id,
    p.name,
    p.created_at,
    p.updated_at,
    MAX(a.occurred_at) as last_action_at
FROM person p
LEFT JOIN action a ON p.id = a.person_id
GROUP BY p.id, p.name, p.created_at, p.updated_at
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;
