package observability

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/tracequery"
	"go.uber.org/zap"
)

// Deps wires production data sources into observability APIs.
type Deps struct {
	Log         *zap.Logger
	Mem         *Store
	Spans       *tracequery.SpanStore
	PG          *PostgresRepo
	Collectors  *CollectorsRepo
	Prom        *PromQLClient
	CH          driver.Conn
	Pool        *pgxpool.Pool
	Identity    *auth.IdentityStore
	SearchURL   string
	SSOManager  *auth.SSOManager
}
