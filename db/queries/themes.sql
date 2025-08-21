-- name: CreateTheme :one
INSERT INTO theme (id, person_id, text)
VALUES (x2b(sqlc.arg(id)), x2b(sqlc.arg(person_id)), sqlc.arg(text))
RETURNING sqlc.embed(theme);

-- name: GetThemeByID :one
SELECT sqlc.embed(theme)
FROM theme
WHERE id = x2b(sqlc.arg(id));

-- name: ListThemes :many
SELECT sqlc.embed(theme)
FROM theme
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListThemesByPersonID :many
SELECT sqlc.embed(theme)
FROM theme
WHERE person_id = x2b(sqlc.arg(person_id))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeleteTheme :exec
DELETE FROM theme
WHERE id = x2b(sqlc.arg(id));
