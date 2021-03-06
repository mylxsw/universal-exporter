package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/universal-exporter/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServiceProvider struct{}

func (s ServiceProvider) Register(app container.Container) {}

func (s ServiceProvider) Boot(app infra.Glacier) {
	app.MustResolve(func(conf *config.Config) {
		app.WebAppRouter(routers(app.Container()))
		app.WebAppMuxRouter(func(router *mux.Router) {
			// prometheus metrics
			router.PathPrefix("/metrics").Handler(promhttp.Handler())
			// health check
			router.PathPrefix("/health").Handler(HealthCheck{})
		})
	})
}

type HealthCheck struct{}

func (h HealthCheck) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(`{"status": "UP"}`))
}
