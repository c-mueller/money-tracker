package service

import "context"

type contextKey string

const userIDKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}
