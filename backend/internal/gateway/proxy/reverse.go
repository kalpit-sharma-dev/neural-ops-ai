package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"go.uber.org/zap"
)

// Handler forwards requests to a backend service.
type Handler struct {
	proxy *httputil.ReverseProxy
	log   *zap.Logger
}

// New creates a reverse proxy handler for a backend base URL.
func New(baseURL string, log *zap.Logger) (*Handler, error) {
	return NewWithTransport(baseURL, log, http.DefaultTransport)
}

// NewWithTransport creates a reverse proxy with a custom upstream transport (e.g. mTLS).
func NewWithTransport(baseURL string, log *zap.Logger, transport http.RoundTripper) (*Handler, error) {
	target, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	if transport != nil {
		proxy.Transport = transport
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("proxy error", zap.String("path", r.URL.Path), zap.Error(err))
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"status":"error","errorCode":"GWY001","message":"upstream unavailable"}`))
	}
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	return &Handler{proxy: proxy, log: log}, nil
}

// ServeHTTP forwards the request using gin context.
func (h *Handler) ServeHTTP(c *gin.Context) {
	if principal, ok := auth.PrincipalFromGin(c); ok {
		c.Request.Header.Set("X-Tenant-ID", principal.TenantID)
		c.Request.Header.Set("X-User-ID", principal.UserID)
		c.Request.Header.Set("X-User-Role", string(principal.Role))
		if len(principal.Services) > 0 {
			c.Request.Header.Set("X-Allowed-Services", strings.Join(principal.Services, ","))
		}
	}
	h.proxy.ServeHTTP(c.Writer, c.Request)
}

// GinHandler returns a gin-compatible handler.
func (h *Handler) GinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c)
	}
}
