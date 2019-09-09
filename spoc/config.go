package spoc

import (
	"github.com/umich-vci/gospoc"
)

// Config holds the provider configuration
type Config struct {
	Username     string
	Password     string
	SPOCEndpoint string
	URLScheme    string
	SSLVerify    bool
}

// Client returns a new client for accessing Spectrum Protect Operations Center
func (c *Config) Client() (*gospoc.Client, error) {

	gospocConfig := new(gospoc.Config)
	gospocConfig.Username = c.Username
	gospocConfig.Password = c.Password
	gospocConfig.URLScheme = c.URLScheme
	gospocConfig.OCHost = c.SPOCEndpoint
	gospocConfig.SSLVerify = c.SSLVerify

	return gospoc.NewClient(gospocConfig)
}
