package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/neuralops/platform/internal/gateway/config"
)

type crewjamSAMLProvider struct {
	sp  *saml.ServiceProvider
	cfg config.SAMLConfig
}

func newCrewjamSAMLProvider(cfg config.SAMLConfig) (*crewjamSAMLProvider, error) {
	if cfg.ACSURL == "" {
		return nil, fmt.Errorf("saml acs url required")
	}
	if cfg.EntityID == "" {
		cfg.EntityID = "neuralops-gateway"
	}

	acsURL, err := url.Parse(cfg.ACSURL)
	if err != nil {
		return nil, fmt.Errorf("parse acs url: %w", err)
	}

	sp := &saml.ServiceProvider{
		EntityID:          cfg.EntityID,
		AcsURL:            *acsURL,
		HTTPClient:        &http.Client{Timeout: 15 * time.Second},
		AllowIDPInitiated: true,
	}

	if cfg.MetadataURL != "" {
		metadataURL, err := url.Parse(cfg.MetadataURL)
		if err != nil {
			return nil, fmt.Errorf("parse metadata url: %w", err)
		}
		metadata, err := samlsp.FetchMetadata(context.Background(), sp.HTTPClient, *metadataURL)
		if err != nil {
			return nil, fmt.Errorf("fetch idp metadata: %w", err)
		}
		sp.IDPMetadata = metadata
	}

	if err := loadServiceProviderKey(sp, cfg); err != nil {
		return nil, err
	}

	return &crewjamSAMLProvider{sp: sp, cfg: cfg}, nil
}

func loadServiceProviderKey(sp *saml.ServiceProvider, cfg config.SAMLConfig) error {
	certPEM := cfg.CertificatePEM
	keyPEM := cfg.KeyPEM
	if certPEM == "" && cfg.CertificateFile != "" {
		raw, err := os.ReadFile(cfg.CertificateFile)
		if err != nil {
			return fmt.Errorf("read saml cert file: %w", err)
		}
		certPEM = string(raw)
	}
	if keyPEM == "" && cfg.KeyFile != "" {
		raw, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return fmt.Errorf("read saml key file: %w", err)
		}
		keyPEM = string(raw)
	}
	if certPEM == "" || keyPEM == "" {
		return nil
	}

	keyPair, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return fmt.Errorf("load saml key pair: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return fmt.Errorf("parse saml leaf cert: %w", err)
	}

	privateKey, ok := keyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("saml key must be rsa")
	}
	sp.Key = privateKey
	sp.Certificate = keyPair.Leaf
	return nil
}

func (p *crewjamSAMLProvider) WriteMetadata(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	descriptor := p.sp.Metadata()
	raw, err := xml.Marshal(descriptor)
	if err != nil {
		http.Error(w, "metadata marshal failed", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(raw)
}

func (p *crewjamSAMLProvider) LoginURL(relayState string) (string, error) {
	idpURL := p.sp.GetSSOBindingLocation(saml.HTTPRedirectBinding)
	req, err := p.sp.MakeAuthenticationRequest(idpURL, saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return "", err
	}
	redirectURL, err := req.Redirect(relayState, p.sp)
	if err != nil {
		return "", err
	}
	return redirectURL.String(), nil
}

func (p *crewjamSAMLProvider) ParseACS(r *http.Request) (SAMLIdentity, error) {
	if err := r.ParseForm(); err != nil {
		return SAMLIdentity{}, err
	}
	assertion, err := p.sp.ParseResponse(r, []string{})
	if err != nil {
		return SAMLIdentity{}, err
	}
	return identityFromAssertion(assertion, p.cfg), nil
}

func identityFromAssertion(assertion *saml.Assertion, cfg config.SAMLConfig) SAMLIdentity {
	identity := SAMLIdentity{
		NameID:   assertion.Subject.NameID.Value,
		TenantID: cfg.DefaultTenant,
		Role:     ParseRole(cfg.DefaultRole),
	}
	for _, statement := range assertion.AttributeStatements {
		for _, attr := range statement.Attributes {
			val := ""
			if len(attr.Values) > 0 {
				val = attr.Values[0].Value
			}
			switch strings.ToLower(attr.Name) {
			case "email", "mail":
				identity.Email = val
			case "tenantid", "tenant_id":
				if val != "" {
					identity.TenantID = val
				}
			case "role":
				if val != "" {
					identity.Role = ParseRole(val)
				}
			}
		}
	}
	if identity.Email == "" {
		identity.Email = identity.NameID
	}
	return identity
}

// ParseACS satisfies SAMLProvider for the dev XML parser.
func (f *SAMLFlow) ParseACS(r *http.Request) (SAMLIdentity, error) {
	if err := r.ParseForm(); err != nil {
		return SAMLIdentity{}, err
	}
	response := r.FormValue("SAMLResponse")
	if response == "" {
		return SAMLIdentity{}, fmt.Errorf("SAMLResponse required")
	}
	return f.ParseACSResponse(response)
}

// WriteMetadata and LoginURL on SAMLFlow implement SAMLProvider.
var _ SAMLProvider = (*SAMLFlow)(nil)
var _ SAMLProvider = (*crewjamSAMLProvider)(nil)

// unused import guard for io
var _ = io.Discard
