package idp

// AuthentikClientConfig authentik manager client configurations.
type AuthentikClientConfig struct {
	Issuer        string
	ClientID      string
	Username      string
	Password      string
	TokenEndpoint string
	GrantType     string
}
