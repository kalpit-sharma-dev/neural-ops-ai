package seed

const (
	DemoTenantID   = "00000000-0000-0000-0000-000000000002"
	DemoTenantName = "NeuralOps Demo"
	DemoEmail      = "demo@neuralops.ai"
	DemoPassword   = "Demo@1234"

	UPIOutageIncidentID  = "00000000-0000-0000-0000-000000000301"
	UPIOutageDeploymentID = "00000000-0000-0000-0000-000000000501"

	DefaultLogCount         = 50_000
	DefaultTransactionCount = 500
	UPIFailedTxnCount       = 1_247
	DefaultDeploymentCount  = 100
)

var DemoServices = []string{
	"api-gateway",
	"auth-service",
	"upi-service",
	"payment-api",
	"ledger-service",
	"notification-service",
	"cbs-adapter",
	"fraud-service",
}

var TxnTypes = []string{"UPI", "NEFT", "RTGS", "IMPS"}
