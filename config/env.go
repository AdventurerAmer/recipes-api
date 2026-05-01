package config

type Environment string

const (
	Dev  Environment = "dev"
	Test Environment = "test"
	Prod Environment = "prod"
)

func (e Environment) String() string {
	return string(e)
}

var Environments = []Environment{
	Dev,
	Test,
	Prod,
}
