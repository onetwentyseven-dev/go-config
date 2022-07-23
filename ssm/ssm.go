package ssm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/onetwentyseven-dev/go-config"
	"github.com/pkg/errors"
)

var _ config.Source = new(Source)

// ParamStore represents the Systems Manager Client methods needed by the ssm config source
type ParamStore interface {
	GetParameters(context.Context, *ssm.GetParametersInput, ...func(*ssm.Options)) (*ssm.GetParametersOutput, error)
}

// Source is a source that pulls parameters from AWS Parameter Store
type Source struct {
	Prefix string
	Ssm    ParamStore
}

// New creates a new source
func New(prefix string, ssmClient ParamStore) *Source {
	return &Source{
		Prefix: prefix,
		Ssm:    ssmClient,
	}
}

// TagKey returns the tag key for the ssm source
func (s *Source) TagKey() string {
	return "ssm"
}

func (s *Source) getParameters(names []string) (map[string]string, error) {
	parameters := make([]types.Parameter, 0, len(names))

	for i := 0; i < len(names); i += 10 {
		end := i + 10
		if end > len(names) {
			end = len(names)
		}

		// TODO: should we get the context from somewhere?
		response, err := s.Ssm.GetParameters(context.Background(), &ssm.GetParametersInput{
			Names:          names[i:end],
			WithDecryption: true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error fetching items from param store")
		}

		parameters = append(parameters, response.Parameters...)
	}

	result := make(map[string]string, len(parameters))
	for _, p := range parameters {
		if p.Name == nil || p.Value == nil {
			continue
		}

		result[*p.Name] = *p.Value
	}

	return result, nil
}

// Process handles processing of ssm configuration parameters
func (s *Source) Process(paramMap map[string][]config.Parameter) error {
	names := make([]string, 0, len(paramMap))
	handlers := make(map[string][]config.Parameter, len(paramMap))

	for name, params := range paramMap {
		name = getParamName(name, s.Prefix)

		names = append(names, name)
		handlers[name] = params
	}

	parameters, err := s.getParameters(names)
	if err != nil {
		return errors.Wrap(err, "error getting parameters")
	}

	for name, params := range handlers {
		val := parameters[name]

		for _, p := range params {
			if err := p.SetValue(val); err != nil {
				return err
			}
		}
	}

	return nil
}

func getParamName(tagValue, prefix string) string {
	tagValue = strings.TrimSpace(strings.ToLower(tagValue))
	parts := strings.Split(tagValue, ",")

	var absolute bool
	if len(parts) > 1 {
		absolute = contains(parts[1:], "absolute")
	}

	result := parts[0]

	if absolute || strings.HasPrefix(result, prefix) {
		return result
	}

	return fmt.Sprintf("%s%s", prefix, result)
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
