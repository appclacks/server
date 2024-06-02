package http

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/appclacks/server/internal/http/handlers"
	"github.com/appclacks/server/internal/http/middlewares"
	"github.com/appclacks/server/internal/validator"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config *Configuration
	server *echo.Echo
	wg     sync.WaitGroup
	logger *slog.Logger
}

type CustomValidator struct {
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := validator.Validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewServer(logger *slog.Logger, config Configuration, registry *prometheus.Registry, builder *handlers.Builder) (*Server, error) {
	err := validator.Validator.Struct(config)
	if err != nil {
		return nil, err
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = &CustomValidator{}
	respCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count the number of HTTP responses.",
		},
		[]string{"method", "status", "path"})

	buckets := []float64{
		0.05, 0.1, 0.2, 0.4, 0.8, 1,
		1.5, 2, 3, 5}
	err = registry.Register(respCounter)
	if err != nil {
		return nil, err
	}

	reqHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_second",
			Help:    "Time to execute http requests",
			Buckets: buckets,
		},
		[]string{"method", "path"})

	err = registry.Register(reqHistogram)
	if err != nil {
		return nil, err
	}

	e.HTTPErrorHandler = errorHandler(logger)
	e.Use(middlewares.MetricsMiddleware(reqHistogram, respCounter, logger))
	e.GET("/healthz", func(ec echo.Context) error {
		return ec.JSON(http.StatusOK, "ok")
	})

	if config.Metrics.BasicAuth.Username != "" {
		basicAuth := echomw.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(config.Metrics.BasicAuth.Username)) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(config.Metrics.BasicAuth.Password)) == 1 {
				return true, nil
			}
			return false, nil
		})
		e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})), basicAuth)
	} else {
		e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	}

	apiGroup := e.Group("/api/v1")
	if config.BasicAuth.Username != "" {
		basicAuth := echomw.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(config.BasicAuth.Username)) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(config.BasicAuth.Password)) == 1 {
				return true, nil
			}
			return false, nil
		})
		apiGroup.Use(basicAuth)
	}

	apiGroup.POST("/healthcheck/dns", builder.CreateDNSHealthcheck)
	apiGroup.PUT("/healthcheck/dns/:id", builder.UpdateDNSHealthcheck)
	apiGroup.POST("/healthcheck/tcp", builder.CreateTCPHealthcheck)
	apiGroup.PUT("/healthcheck/tcp/:id", builder.UpdateTCPHealthcheck)
	apiGroup.POST("/healthcheck/http", builder.CreateHTTPHealthcheck)
	apiGroup.PUT("/healthcheck/http/:id", builder.UpdateHTTPHealthcheck)
	apiGroup.POST("/healthcheck/tls", builder.CreateTLSHealthcheck)
	apiGroup.PUT("/healthcheck/tls/:id", builder.UpdateTLSHealthcheck)
	apiGroup.POST("/healthcheck/command", builder.CreateCommandHealthcheck)
	apiGroup.PUT("/healthcheck/command/:id", builder.UpdateCommandHealthcheck)
	apiGroup.DELETE("/healthcheck/:id", builder.DeleteHealthcheck)
	apiGroup.GET("/healthcheck/:identifier", builder.GetHealthcheck)
	apiGroup.GET("/healthcheck", builder.ListHealthchecks)
	apiGroup.GET("/cabourotte/discovery", builder.CabourotteDiscovery)

	return &Server{
		server: e,
		config: &config,
		logger: logger,
	}, nil

}

func (s *Server) Start() error {
	address := fmt.Sprintf("[%s]:%d", s.config.Host, s.config.Port)
	s.logger.Info(fmt.Sprintf("http server starting on %s", address))
	if s.config.Cert != "" {
		s.logger.Info("tls is enabled on the http server")
		tlsConfig, err := getTLSConfig(s.config.Key, s.config.Cert, s.config.Cacert, s.config.ServerName, s.config.Insecure)
		if err != nil {
			return err
		}

		s.server.TLSServer.TLSConfig = tlsConfig
		tlsServer := s.server.TLSServer
		tlsServer.Addr = address
		if !s.server.DisableHTTP2 {
			tlsServer.TLSConfig.NextProtos = append(tlsServer.TLSConfig.NextProtos, "h2")
		}
	}

	go func() {
		defer s.wg.Done()
		var err error
		if s.config.Cert != "" {
			err = s.server.StartServer(s.server.TLSServer)
		} else {
			err = s.server.Start(address)

		}
		if err != http.ErrServerClosed {
			s.logger.Error(fmt.Sprintf("http server error: %s", err.Error()))
			os.Exit(2)
		}

	}()
	s.wg.Add(1)
	return nil
}

func (s *Server) Stop() error {
	s.logger.Info("stopping the http server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	s.wg.Wait()
	if err != nil {
		return err
	}
	s.logger.Info("http server stopped")
	return nil
}
