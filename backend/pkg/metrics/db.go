package metrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var dbPoolOpenConnections = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "db_pool_open_connections",
		Help: "Open connections in a database pool",
	},
	[]string{"service", "pool"},
)

var dbPoolInUseConnections = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "db_pool_in_use_connections",
		Help: "In-use connections in a database pool",
	},
	[]string{"service", "pool"},
)

func init() {
	prometheus.MustRegister(dbPoolOpenConnections, dbPoolInUseConnections)
}

// ObserveDBPool publishes pgx pool stats for Prometheus scraping.
func ObserveDBPool(service, poolName string, stats *pgxpool.Stat) {
	if stats == nil {
		return
	}
	dbPoolOpenConnections.WithLabelValues(service, poolName).Set(float64(stats.TotalConns()))
	dbPoolInUseConnections.WithLabelValues(service, poolName).Set(float64(stats.AcquiredConns()))
}

// StartDBPoolReporter periodically publishes pool stats until context cancellation.
func StartDBPoolReporter(ctx context.Context, service, poolName string, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ObserveDBPool(service, poolName, pool.Stat())
			}
		}
	}()
}
