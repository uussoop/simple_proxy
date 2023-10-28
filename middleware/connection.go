package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

func SetOpenAIServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(utils.UserKey).(*database.User)

		var leastBusyEndpoint database.Endpoint
		var load int = 10000

		ens, err := u.GetEndpoints()

		if err != nil || len(ens) == 0 {
			logrus.Error("error getting endpoints for user <", u.Name, ">", err)
			api.InternalServerError(w)
			return
		}

		logrus.Infof("%+v", r)

		bodyCopy, readErr := io.ReadAll(r.Body)

		if readErr != nil {
			logrus.Errorf("error reading body: %s\n", readErr)
			api.InternalServerError(w)
			return
		}

		type body struct {
			Model string `json:"model"`
		}

		responseBody := &body{}

		logrus.Info(bodyCopy)

		err = json.Unmarshal(bodyCopy, &responseBody)

		r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))

		if err != nil {
			logrus.Errorf("couldn't unmarshal the model: %s\n", err)
			next.ServeHTTP(w, r)
			return
		}

		if responseBody.Model == "" {
			api.BadRequest(w, "model is required")
			return
		}

		m := database.Model{}

		err = m.Get(responseBody.Model)

		if err != nil {
			api.BadRequest(w, "model not found")
			return
		}

		if !u.HasModel(m) {
			api.BadRequest(w, "user doesn't have access to this model")
			return
		}

		for _, en := range ens {
			rim, _ := en.GetRequestInMin()
			rid, _ := en.GetRequestInDay()
			if en.GetConnection() < load && en.GetConnection() < en.Concurrent && en.Active() && rid < en.RPD && rim < en.RPM && en.HasModel(m) {
				load = en.Connections
				leastBusyEndpoint = en
			}
		}

		if leastBusyEndpoint.ID == 0 {
			api.ServerNotReady(w)
			return
		}

		leastBusyEndpoint.AddConnection()

		c := context.WithValue(r.Context(), utils.EndpointKey, &leastBusyEndpoint)
		c = context.WithValue(c, utils.ModelKey, &m)

		customWriter := &utils.StatusCaptureResponseWriter{ResponseWriter: w}

		logrus.Debug("user <", u.Name, "> is sending request to endpoint <", leastBusyEndpoint.Name, ">")

		use := database.EndpointModelUsage{}

		use.GetOrCreate(*u, leastBusyEndpoint, m)

		next.ServeHTTP(customWriter, r.Clone(c))
		if customWriter.StatusCode >= 200 && customWriter.StatusCode < 300 {
			leastBusyEndpoint.Requested()
		}
		leastBusyEndpoint.RemoveConnection()
		logrus.Debug("user <", u.Name, "> is done sending request to endpoint <", leastBusyEndpoint.Name, "> Current Connections:", leastBusyEndpoint.Connections)
	})
}
