-- name: CreateTheme :one
INSERT INTO theme (id, person_id, text)
VALUES (x2b($1), x2b($2), $3)
RETURNING sqlc.embed(theme);

-- name: GetThemeByID :one
SELECT sqlc.embed(theme)
FROM theme
WHERE id = x2b($1);

-- name: ListThemes :many
SELECT sqlc.embed(theme)
FROM theme
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListThemesByPersonID :many
SELECT sqlc.embed(theme)
FROM theme
WHERE person_id = x2b($1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteTheme :exec
DELETE FROM theme
WHERE id = x2b($1);
