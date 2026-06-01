package auth

import (
	"encoding/base64"
	"testing"

	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/stretchr/testify/require"
)

func TestSAMLFlowMetadataAndParse(t *testing.T) {
	flow, err := NewSAMLFlow(config.SAMLConfig{
		Enabled:       true,
		EntityID:      "neuralops-sp",
		ACSURL:        "http://localhost:8080/api/v1/auth/saml/acs",
		SSOURL:        "http://idp.example/sso",
		DefaultTenant: "tenant-1",
		DefaultRole:   "SRE",
	})
	require.NoError(t, err)
	require.Contains(t, flow.MetadataXML(), "neuralops-sp")

	loginURL, err := flow.LoginURL("relay")
	require.NoError(t, err)
	require.Contains(t, loginURL, "http://idp.example/sso")

	raw := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
	<Assertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion">
		<Subject><NameID>demo@neuralops.ai</NameID></Subject>
		<AttributeStatement>
			<Attribute Name="email"><AttributeValue>demo@neuralops.ai</AttributeValue></Attribute>
		</AttributeStatement>
	</Assertion>
</samlp:Response>`
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	identity, err := flow.ParseACSResponse(encoded)
	require.NoError(t, err)
	require.Equal(t, "demo@neuralops.ai", identity.Email)
	require.Equal(t, RoleSRE, identity.Role)
}
