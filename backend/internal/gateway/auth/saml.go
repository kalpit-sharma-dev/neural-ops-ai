package auth

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/neuralops/platform/internal/gateway/config"
)

// SAMLFlow handles SAML SP metadata and ACS for enterprise IdP integration.
type SAMLFlow struct {
	cfg config.SAMLConfig
}

// NewSAMLFlow creates a SAML service provider flow when enabled.
func NewSAMLFlow(cfg config.SAMLConfig) (*SAMLFlow, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.EntityID == "" {
		cfg.EntityID = "neuralops-gateway"
	}
	if cfg.ACSURL == "" {
		return nil, fmt.Errorf("saml acs url required")
	}
	return &SAMLFlow{cfg: cfg}, nil
}

// MetadataXML returns SP metadata for IdP federation.
func (f *SAMLFlow) MetadataXML() string {
	return fmt.Sprintf(`<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <SPSSODescriptor AuthnRequestsSigned="false" WantAssertionsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s" index="1"/>
  </SPSSODescriptor>
</EntityDescriptor>`, xmlEscape(f.cfg.EntityID), xmlEscape(f.cfg.ACSURL))
}

// LoginURL returns the IdP SSO redirect URL.
func (f *SAMLFlow) LoginURL(relayState string) (string, error) {
	if f.cfg.SSOURL == "" {
		return "", fmt.Errorf("saml sso url not configured")
	}
	values := url.Values{}
	values.Set("SAMLRequest", base64.StdEncoding.EncodeToString([]byte("NeuralOpsAuthnRequest")))
	if relayState != "" {
		values.Set("RelayState", relayState)
	}
	sep := "?"
	if strings.Contains(f.cfg.SSOURL, "?") {
		sep = "&"
	}
	return f.cfg.SSOURL + sep + values.Encode(), nil
}

// SAMLIdentity is extracted from a SAML assertion response.
type SAMLIdentity struct {
	Email    string
	NameID   string
	TenantID string
	Role     Role
}

type samlResponse struct {
	Assertion struct {
		Subject struct {
			NameID string `xml:"NameID"`
		} `xml:"Subject"`
		AttributeStatement struct {
			Attributes []struct {
				Name   string `xml:"Name,attr"`
				Values []struct {
					Value string `xml:"Value"`
				} `xml:"AttributeValue"`
			} `xml:"Attribute"`
		} `xml:"AttributeStatement"`
	} `xml:"Assertion"`
}

// ParseACSResponse decodes a SAMLResponse POST body (unsigned dev/compose mode).
func (f *SAMLFlow) ParseACSResponse(samlResponseB64 string) (SAMLIdentity, error) {
	raw, err := base64.StdEncoding.DecodeString(samlResponseB64)
	if err != nil {
		return SAMLIdentity{}, fmt.Errorf("decode saml response: %w", err)
	}

	var parsed samlResponse
	if err := xml.Unmarshal(raw, &parsed); err != nil {
		return SAMLIdentity{}, fmt.Errorf("parse saml response: %w", err)
	}

	identity := SAMLIdentity{
		NameID:   strings.TrimSpace(parsed.Assertion.Subject.NameID),
		TenantID: f.cfg.DefaultTenant,
		Role:     ParseRole(f.cfg.DefaultRole),
	}
	for _, attr := range parsed.Assertion.AttributeStatement.Attributes {
		value := ""
		if len(attr.Values) > 0 {
			value = strings.TrimSpace(attr.Values[0].Value)
		}
		switch strings.ToLower(attr.Name) {
		case "email", "mail", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
			identity.Email = value
		case "tenantid", "tenant_id":
			if value != "" {
				identity.TenantID = value
			}
		case "role":
			if value != "" {
				identity.Role = ParseRole(value)
			}
		}
	}
	if identity.Email == "" {
		identity.Email = identity.NameID
	}
	if identity.Email == "" {
		return SAMLIdentity{}, fmt.Errorf("saml assertion missing identity")
	}
	return identity, nil
}

// WriteMetadata serves SP metadata XML.
func (f *SAMLFlow) WriteMetadata(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	_, _ = w.Write([]byte(f.MetadataXML()))
}

func xmlEscape(value string) string {
	replacer := strings.NewReplacer(`&`, "&amp;", `"`, "&quot;", `'`, "&apos;", `<`, "&lt;", `>`, "&gt;")
	return replacer.Replace(value)
}
