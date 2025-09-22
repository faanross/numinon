package listener

import (
	"context"
	"net/http"
)

// WithListenerID returns middleware that adds the listener ID to request context
func WithListenerID(listenerID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add listener ID to context
		ctx := context.WithValue(r.Context(), "listenerID", listenerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
