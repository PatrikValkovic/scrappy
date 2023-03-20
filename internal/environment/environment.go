package environment

import "errors"

const (
	Development = "development"
	Production  = "production"
)

func ValidateEnvironment(env string) error {
	if env != Development && env != Production {
		return errors.New("Invalid environment")
	}
	return nil
}
