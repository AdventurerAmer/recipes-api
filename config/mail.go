package config

type Mailer struct {
	From     string `koanf:"from" validate:"required,max=128"`
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     *int   `koanf:"port" validate:"required,max=65535"`
	Username string `kaonf:"username" validate:"required,max=128"`
	Password string `kaonf:"username" validate:"required,max=128"`
}
