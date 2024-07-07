package http

type BasicAuth struct {
	Username string
	Password string
}

type Metrics struct {
	BasicAuth BasicAuth `yaml:"basic-auth"`
}

type Configuration struct {
	Host       string `validate:"required"`
	Port       uint32 `validate:"required"`
	Key        string
	Cert       string
	Cacert     string
	Insecure   bool
	ServerName string    `yaml:"server-name"`
	BasicAuth  BasicAuth `yaml:"basic-auth"`
	Metrics    Metrics
}
