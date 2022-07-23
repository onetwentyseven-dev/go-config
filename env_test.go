package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvSource_TagKey(t *testing.T) {
	src := &EnvSource{}

	assert.Equal(t, "env", src.TagKey())
}

func TestEnvSource_Process(t *testing.T) {
	testCases := []struct {
		name      string
		src       EnvSource
		vars      map[string]string
		params    map[string]*mockParameter
		expectErr bool
	}{{
		name: "SetValueError",
		src:  EnvSource{},
		vars: map[string]string{
			"TESTING_VAR": "test",
		},
		params: map[string]*mockParameter{
			"TESTING_VAR": {
				setValErr:   errors.New("test error"),
				expectValue: true,
			},
		},
		expectErr: true,
	}, {
		name: "NoValueError",
		src: EnvSource{
			Prefix: "TEST_PREFIX_",
		},
		vars: map[string]string{
			"TESTING_VAR": "test",
		},
		params: map[string]*mockParameter{
			"TESTING_VAR": {
				noValErr:    errors.New("test error"),
				expectValue: false,
			},
		},
		expectErr: true,
	}, {
		name: "Normal",
		src:  EnvSource{},
		vars: map[string]string{
			"TESTING_VAR": "test",
		},
		params: map[string]*mockParameter{
			"TESTING_VAR": {
				expectValue: true,
			},
		},
		expectErr: false,
	}, {
		name: "Prefixed",
		src: EnvSource{
			Prefix: "TEST_PREFIX_",
		},
		vars: map[string]string{
			"TEST_PREFIX_TESTING_VAR": "test",
			"TESTING_VAR2":            "test2",
		},
		params: map[string]*mockParameter{
			"TESTING_VAR": {
				expectValue: true,
			},
			"TESTING_VAR2": {
				expectValue: false,
			},
		},
		expectErr: false,
	}, {
		name: "StrictCase",
		src: EnvSource{
			StrictCase: true,
		},
		vars: map[string]string{
			"testing_var":  "test",
			"TESTING_VAR2": "test2",
		},
		params: map[string]*mockParameter{
			"testing_var": {
				expectValue: true,
			},
			"testing_var2": {
				expectValue: false,
			},
		},
		expectErr: false,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tc.vars {
				assert.NoError(t, os.Setenv(k, v))
			}

			params := make(map[string][]Parameter, len(tc.params))
			for k, p := range tc.params {
				params[k] = []Parameter{p}
			}

			err := tc.src.Process(params)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, p := range tc.params {
				p.AssertExpectations(t)
			}
		})
	}
}
