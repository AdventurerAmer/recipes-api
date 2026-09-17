package config

type App struct {
	Name    string `koanf:"name" validate:"required,max=128"`
	Domain  string `koanf:"domain" validate:"required,fqdn"`
	Version string `koanf:"version" validate:"required,semver"`
}
