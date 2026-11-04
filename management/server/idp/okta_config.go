package idp

// OktaClientConfig okta manager client configurations.
type OktaClientConfig struct {
	APIToken      string
	Issuer        string
	TokenEndpoint string
	GrantType     string
}
