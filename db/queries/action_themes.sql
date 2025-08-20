-- name: AddThemeToAction :exec
INSERT INTO action_themes (action_id, theme_id)
VALUES (x2b($1), x2b($2));

-- name: RemoveThemeFromAction :exec
DELETE FROM action_themes
WHERE action_id = x2b($1) AND theme_id = x2b($2);

-- name: ListThemesByActionID :many
SELECT sqlc.embed(theme)
FROM action_themes at
JOIN theme ON at.theme_id = theme.id
WHERE at.action_id = x2b($1)
ORDER BY theme.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActionsByThemeID :many
SELECT sqlc.embed(action)
FROM action_themes at
JOIN action ON at.action_id = action.id
WHERE at.theme_id = x2b($1)
ORDER BY action.created_at DESC
LIMIT $2 OFFSET $3;
