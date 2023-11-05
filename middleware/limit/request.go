package limit

import (
	"net/http"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

func LimitRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(utils.UserKey).(*database.User)

		if u.Requested() > u.RateLimit {
			logrus.Debug("user <", u.Name, "> is limited in sending requests")
			api.RateLimitError(w)
			return
		}

		// Create a custom ResponseWriter to capture the status code
		customWriter := &utils.StatusCaptureResponseWriter{ResponseWriter: w}

		next.ServeHTTP(customWriter, r)

		// Access the captured status code
		statusCode := customWriter.StatusCode

		logrus.Info("limit request Status Code:", statusCode)

		// You can now use the statusCode as needed
		if !(customWriter.StatusCode >= 200 && customWriter.StatusCode < 300) {
			u.RemoveRequested()
		}
	})
}

// Create a custom ResponseWriter to capture the status code
