-- name: AddThemeToConversation :exec
INSERT INTO conversation_theme (conversation_id, theme_id)
VALUES (x2b(sqlc.arg(conversation_id)), x2b(sqlc.arg(theme_id)));
