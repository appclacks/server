package http

// BasicAuth basic auth for the configuration
type BasicAuth struct {
	Username string
	Password string
}

// Configuration the HTTP server configuration
type Configuration struct {
	Host       string `validate:"required"`
	Port       uint32 `validate:"required"`
	Key        string
	Cert       string
	Cacert     string
	Insecure   bool
	ServerName string
	BasicAuth  BasicAuth `yaml:"basic-auth"`
}
