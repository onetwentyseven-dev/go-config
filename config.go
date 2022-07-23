// Package config implements an envconfig-like interface that can load values the environment, as well as additional data sources
package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// Source represents a configuration source
type Source interface {
	// TagKey returns the key used for this particular source
	TagKey() string

	// Process processes a given set of parameters, returning an error if one occurred
	Process(map[string][]Parameter) error
}

func process(params interface{}, paramMap map[string]map[string][]Parameter, sourceKeys []string) error {
	var errs *multierror.Error

	typeOf := reflect.TypeOf(params)
	elem := reflect.ValueOf(params).Elem()

	numFields := elem.NumField()
	for i := 0; i < numFields; i++ {
		field := elem.Field(i)

		// handle interface
		if field.Kind() == reflect.Interface && !field.IsNil() {
			if err := process(field.Interface(), paramMap, sourceKeys); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error processing interface: %w", err))
			}

			// interfaces are handled recursively, continue
			continue
		}

		if field.Kind() == reflect.Struct {
			val := field.Addr()
			if err := process(val.Interface(), paramMap, sourceKeys); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error processing struct: %w", err))
			}

			// structs are handled recursively, continue
			continue
		}

		sf := typeOf.Elem().Field(i)
		ignoreValue, ok := sf.Tag.Lookup(ignoreTag)
		if ok {
			ignoreValueBool, err := strconv.ParseBool(ignoreValue)
			if err == nil {
				if ignoreValueBool {
					continue
				}
			}
		}

		setFn, err := getSetter(field)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error getting set function: %w", err))
			continue
		}

		var foundHandler bool

		// iterate through source tag keys and populate parameter map with parameter
		// for any found tags
		for tagKey, sourceParams := range paramMap {
			tagValue, ok := sf.Tag.Lookup(tagKey)
			if !ok {
				continue
			}

			foundHandler = true

			param := newParameter(sf, setFn, tagKey, tagValue)
			sourceParams[tagValue] = append(sourceParams[tagValue], param)

			// since map ordering is non-deterministic, the docs will call out potential
			// strange behavior if tags from multiple sources are specified on the same field.
			//
			// to prevent unnecessary lookups in a given remote source, we'll just assume there's
			// only one source tag on a parameter and break after finding one
			break
		}

		if !foundHandler {
			required, _ := strconv.ParseBool(sf.Tag.Get(requiredTag))
			if !required {
				continue
			}

			errs = multierror.Append(
				errs,
				fmt.Errorf(
					"error: the field %s was marked as required, but did not specify a struct tag for one of the provided sources: %s",
					sf.Name,
					strings.Join(sourceKeys, ", "),
				),
			)
		}
	}

	return errs.ErrorOrNil()
}

func processSources(sources []Source, paramMap map[string]map[string][]Parameter) error {
	var errs *multierror.Error

	for _, s := range sources {
		params, ok := paramMap[s.TagKey()]
		if !ok || len(params) == 0 {
			continue
		}

		errs = multierror.Append(errs, s.Process(params))
	}

	return errs.ErrorOrNil()
}

// Process handles processing values from various sources
func Process(params interface{}, sources ...Source) error {
	hasEnvSource := false
	for _, s := range sources {
		if _, ok := s.(*EnvSource); ok {
			hasEnvSource = true
			break
		}
	}

	// if no EnvSource exists, add in one with default settings
	if !hasEnvSource {
		sources = append(sources, &EnvSource{})
	}

	paramMap := make(map[string]map[string][]Parameter)
	sourceKeys := make([]string, len(sources))

	for i, s := range sources {
		key := s.TagKey()

		if _, ok := paramMap[key]; ok {
			return fmt.Errorf("error: multiple sources provided with the tag key %s", key)
		}

		paramMap[key] = make(map[string][]Parameter)
		sourceKeys[i] = key
	}

	if err := process(params, paramMap, sourceKeys); err != nil {
		return err
	}

	return processSources(sources, paramMap)
}
