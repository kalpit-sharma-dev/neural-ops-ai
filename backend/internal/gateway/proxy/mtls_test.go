package proxy_test

import (
	"net/http"
	"testing"

	"github.com/neuralops/platform/internal/gateway/proxy"
	"github.com/stretchr/testify/require"
)

func TestBuildTransportDisabled(t *testing.T) {
	t.Parallel()
	rt, err := proxy.BuildTransport(proxy.MTLSConfig{Enabled: false})
	require.NoError(t, err)
	require.Equal(t, http.DefaultTransport, rt)
}

func TestBuildTransportMissingCert(t *testing.T) {
	t.Parallel()
	_, err := proxy.BuildTransport(proxy.MTLSConfig{Enabled: true})
	require.Error(t, err)
}
