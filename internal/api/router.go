package api

import (
	"b3challenge/internal/api/ctrl"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	router *echo.Echo
}

func NewServer() *Server {
	router := echo.New()
	router.HideBanner = true
	router.Use(echomiddleware.LoggerWithConfig(echomiddleware.LoggerConfig{})) //nolint:exhaustruct

	return &Server{
		router: router,
	}
}

func (s *Server) Start(port uint16) {
	const timeout = 15 * time.Second

	srv := &http.Server{ //nolint:exhaustruct
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      s.router,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
	}

	if err := s.router.Start(srv.Addr); err != nil {
		s.router.Logger.Fatal("Failed to start server: ", err)
	}
}

func (s *Server) ConfigureRoutes(tradeCtrl *ctrl.TradesCtrl) {
	s.router.GET("/ticker-metrics", tradeCtrl.ComputeTickerMetrics)
}
