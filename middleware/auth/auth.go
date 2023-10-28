package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

func IsAuthroized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			api.InvalidApiKeyError(w)
			return
		}

		users, exists := database.Authenticate(&token)

		if exists {
			logrus.Debug("user <", users[0].Name, "> is sending request")
			users[0].SetLastSeen(time.Now())

			c := context.WithValue(r.Context(), utils.UserKey, &users[0])

			next.ServeHTTP(w, r.Clone(c))
		}
	})
}
