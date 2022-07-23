package config

import (
	"fmt"
	"os"
	"strings"
)

// EnvSource is a source that loads configuration parameters from environment variables
type EnvSource struct {
	// Optional prefix
	Prefix string
	// Since the convention for environment variables is to have them all be uppercase, i.e. MY_ENVIRONMENT_VARIABLE,
	// by default this source will call ToUpper on each key it tries to load. If you with to disable that behavior and
	// respect the casing present in each property, set this to true
	StrictCase bool
}

// TagKey returns the tag key for the env loader
func (e *EnvSource) TagKey() string {
	return "env"
}

func (e *EnvSource) getEnvVar(key string) string {
	if e.Prefix != "" {
		key = fmt.Sprintf("%s%s", e.Prefix, key)
	}

	if !e.StrictCase {
		key = strings.ToUpper(key)
	}

	return os.Getenv(key)
}

// Process processes values from environment variables
func (e *EnvSource) Process(paramMap map[string][]Parameter) error {
	for key, params := range paramMap {
		val := e.getEnvVar(key)

		for _, param := range params {
			if err := param.SetValue(val); err != nil {
				return err
			}
		}
	}

	return nil
}
