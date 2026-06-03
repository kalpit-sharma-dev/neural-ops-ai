package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/pkg/license"
)

var (
	licenseOnce   sync.Once
	licenseDoc    license.Document
	licenseErr    error
	licenseLoaded time.Time
)

// LicenseEnforcement validates offline license when NEURALOPS_LICENSE_PATH is set.
func LicenseEnforcement() gin.HandlerFunc {
	path := strings.TrimSpace(os.Getenv("NEURALOPS_LICENSE_PATH"))
	secret := strings.TrimSpace(os.Getenv("NEURALOPS_LICENSE_SECRET"))
	if path == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if isLicenseExemptPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		doc, err := cachedLicense(path, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"status":    "error",
				"errorCode": "LIC001",
				"message":   err.Error(),
			})
			return
		}
		c.Set("licenseCustomerId", doc.CustomerID)
		c.Next()
	}
}

func cachedLicense(path, secret string) (license.Document, error) {
	licenseOnce.Do(func() {
		licenseDoc, licenseErr = license.ValidateFile(path, secret)
		licenseLoaded = time.Now().UTC()
	})
	if licenseErr == nil && time.Since(licenseLoaded) > time.Hour {
		licenseDoc, licenseErr = license.ValidateFile(path, secret)
		licenseLoaded = time.Now().UTC()
	}
	return licenseDoc, licenseErr
}

func isLicenseExemptPath(path string) bool {
	switch path {
	case "/health", "/live", "/ready", "/metrics":
		return true
	}
	return strings.HasPrefix(path, "/swagger")
}
