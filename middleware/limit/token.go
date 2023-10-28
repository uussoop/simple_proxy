package limit

import (
	"net/http"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
)

func LimitToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(utils.UserKey).(*database.User)

		if u.IsLimited() {
			api.RateLimitError(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
