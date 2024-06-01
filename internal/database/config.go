package database

type Configuration struct {
	Migrations string `validate:"required"`
	Username   string `validate:"required"`
	Password   string `validate:"required"`
	Database   string `validate:"required"`
	Host       string `validate:"required"`
	Port       uint   `validate:"required,gte=0"`
	SSLMode    string `yaml:"ssl-mode"`
}
