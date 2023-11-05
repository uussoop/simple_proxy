package limit

import (
	"net/http"
	"sync"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

var requestLimitLock sync.Mutex

func LimitRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(utils.UserKey).(*database.User)

		requestLimitLock.Lock()

		if u.IsRateLimited() {
			logrus.Debug("user <", u.Name, "> is limited in sending requests")
			api.RateLimitError(w)
			requestLimitLock.Unlock()
			return
		}

		// Create a custom ResponseWriter to capture the status code
		customWriter := &utils.StatusCaptureResponseWriter{ResponseWriter: w}
		u.Requested()

		requestLimitLock.Unlock()

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
