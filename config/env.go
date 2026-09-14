package config

type Env string

const (
	EnvLocal      Env = "local"
	EnvStaging    Env = "staging"
	EnvProduction Env = "production"
)

func (e Env) String() string {
	return string(e)
}

var EnvValues = []Env{EnvLocal, EnvStaging, EnvProduction}
