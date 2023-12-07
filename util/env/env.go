package env

import "os"

const (
	ENV_PROD  = "prod"
	ENV_UAT   = "uat"
	ENV_DEV   = "dev"
	ENV_TEST  = "test"
	ENV_LOCAL = "local"
)

type Env interface {
	GetEnv() string
	IsEnv(env string) bool
}

type env struct{}

func (e *env) GetEnv() string {
	switch os.Getenv("ENV") {
	case "prod":
		return ENV_PROD
	case "uat":
		return ENV_UAT
	case "test":
		return ENV_TEST
	case "local":
		return ENV_LOCAL
	default:
		return ENV_DEV
	}
}

func (e *env) IsEnv(env string) bool {
	return e.GetEnv() == env
}

func New() *env {
	return &env{}
}
