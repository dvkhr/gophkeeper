package auth

import "context"

// Используется для безопасной передачи идентификатора пользователя в контексте.
// userIDKey — приватный тип ключа для хранения идентификатора пользователя в контексте.
type userIDKey struct{}

// WithUserID добавляет идентификатор пользователя в контекст.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// GetUserID извлекает идентификатор пользователя из контекста.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}
