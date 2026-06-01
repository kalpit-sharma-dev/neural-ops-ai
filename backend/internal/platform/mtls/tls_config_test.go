package mtls_test

import (
	"os"
	"testing"

	"github.com/neuralops/platform/internal/platform/mtls"
	"github.com/stretchr/testify/require"
)

func TestBuildServerTLSRequiresClientCert(t *testing.T) {
	t.Parallel()
	certDir := "../../../infra/certs"
	if _, err := os.Stat(certDir + "/ca.crt"); err != nil {
		t.Skip("dev mTLS certs not generated; run scripts/gen-mtls-certs.sh")
	}
	cfg := mtls.ServerConfig{
		CertFile:          certDir + "/internal-server.crt",
		KeyFile:           certDir + "/internal-server.key",
		CAFile:            certDir + "/ca.crt",
		RequireClientCert: true,
	}
	tlsCfg, err := mtls.BuildServerTLS(cfg)
	require.NoError(t, err)
	require.Equal(t, true, mtls.VerifyPeerRequiresClientCert(cfg))
	require.NotNil(t, tlsCfg.ClientCAs)
}
