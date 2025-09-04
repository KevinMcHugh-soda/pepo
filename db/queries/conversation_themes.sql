-- name: AddThemeToConversation :exec
INSERT INTO conversation_theme (conversation_id, theme_id)
VALUES (x2b(sqlc.arg(conversation_id)), x2b(sqlc.arg(theme_id)));

-- name: ListThemesByConversationID :many
SELECT sqlc.embed(theme)
FROM conversation_theme ct
JOIN theme ON ct.theme_id = theme.id
WHERE ct.conversation_id = x2b(sqlc.arg(conversation_id))
ORDER BY theme.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeleteThemesByConversationID :exec
DELETE FROM conversation_theme
WHERE conversation_id = x2b(sqlc.arg(conversation_id));
