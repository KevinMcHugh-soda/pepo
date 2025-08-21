-- name: AddThemeToAction :exec
INSERT INTO action_theme (action_id, theme_id)
VALUES (x2b(sqlc.arg(action_id)), x2b(sqlc.arg(theme_id)));

-- name: RemoveThemeFromAction :exec
DELETE FROM action_theme
WHERE action_id = x2b(sqlc.arg(action_id)) AND theme_id = x2b(sqlc.arg(theme_id));

-- name: ListThemesByActionID :many
SELECT sqlc.embed(theme)
FROM action_theme at
JOIN theme ON at.theme_id = theme.id
WHERE at.action_id = x2b(sqlc.arg(action_id))
ORDER BY theme.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListActionsByThemeID :many
SELECT sqlc.embed(action)
FROM action_theme at
JOIN action ON at.action_id = action.id
WHERE at.theme_id = x2b(sqlc.arg(theme_id))
ORDER BY action.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
