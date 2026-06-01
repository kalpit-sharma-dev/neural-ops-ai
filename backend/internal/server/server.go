package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/pkg/config"
	"github.com/neuralops/platform/pkg/metrics"
	"go.uber.org/zap"
)

// Run starts the HTTP server with health, readiness, and liveness endpoints.
func Run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(log))

	metrics.RegisterRoutes(router, cfg.ServiceName)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"service": cfg.ServiceName,
			"data": gin.H{
				"healthy": true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"service": cfg.ServiceName,
			"data": gin.H{
				"ready": true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"service": cfg.ServiceName,
			"data": gin.H{
				"alive": true,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	router.GET("/api/v1/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"service":     cfg.ServiceName,
				"environment": cfg.Environment,
				"version":     "0.1.0-phase0",
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown failed", zap.Error(err))
		}
	}()

	log.Info("starting service",
		zap.String("service", cfg.ServiceName),
		zap.Int("port", cfg.HTTPPort),
		zap.String("environment", cfg.Environment),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func requestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
