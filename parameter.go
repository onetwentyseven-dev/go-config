package config

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	defaultTag  = "default"
	requiredTag = "required"
	ignoreTag   = "ignore"
)

// Parameter represents an individual parameter, used for handling by remote sources
type Parameter interface {
	NoValue() error
	SetValue(string) error
}

type parameter struct {
	fieldName        string
	tagKey, tagValue string

	required     bool
	defaultValue string
	setFn        setter
}

func newParameter(sf reflect.StructField, setFn setter, tagKey, tagValue string) Parameter {
	// ignore error parsing required tag, if it's not valid we just assume not required
	required, _ := strconv.ParseBool(sf.Tag.Get(requiredTag))

	return &parameter{
		fieldName:    sf.Name,
		tagKey:       tagKey,
		tagValue:     tagValue,
		required:     required,
		defaultValue: sf.Tag.Get(defaultTag),
		setFn:        setFn,
	}
}

// NoValue handles the case where a value is not found in a source
func (p parameter) NoValue() error {
	if p.required && p.defaultValue == "" {
		return fmt.Errorf(
			"error: field %s was specified as required, but was not found via key %s in source %s",
			p.fieldName,
			p.tagValue,
			p.tagKey,
		)
	}

	if p.defaultValue != "" {
		return p.set(p.defaultValue)
	}

	return nil
}

// SetValue sets a value in the parameter using the value from a source
func (p *parameter) SetValue(val string) error {
	if val == "" {
		return p.NoValue()
	}

	return p.set(val)
}

func (p *parameter) set(val string) error {
	if err := p.setFn(val); err != nil {
		return fmt.Errorf("error setting value for field %s: %w", p.fieldName, err)
	}

	return nil
}
