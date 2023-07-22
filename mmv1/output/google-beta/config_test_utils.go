package google

import (
	"net/http/httptest"
	"strings"
)

// NewTestConfig create a config using the http test server.
func NewTestConfig(server *httptest.Server) *Config {
	cfg := &Config{}
	cfg.Client = server.Client()
	configureTestBasePaths(cfg, server.URL)
	return cfg
}

func configureTestBasePaths(c *Config, url string) {
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	// Generated Products
	c.ComputeBasePath = url

	// Handwritten Products / Versioned / Atypical Entries
	c.CloudBillingBasePath = url
	c.ComposerBasePath = url
	c.ContainerBasePath = url
	c.DataprocBasePath = url
	c.DataflowBasePath = url
	c.IamCredentialsBasePath = url
	c.ResourceManagerV3BasePath = url
	c.IAMBasePath = url
	c.ServiceNetworkingBasePath = url
	c.BigQueryBasePath = url
	c.StorageTransferBasePath = url
	c.BigtableAdminBasePath = url
}
