package config

import (
	"os"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	testCases := []struct {
		name       string
		params     interface{}
		envSource  *EnvSource
		envvars    map[string]string
		mockSource *mockSource

		expectedData map[string]interface{}
		expectErr    bool
	}{{
		name: "DefaultEnvSource",
		params: &struct {
			TestDefaultStr  string        `env:"TEST_DEFAULT_STR" default:"test string" required:"true"`
			TestDefaultInt  int           `env:"TEST_DEFAULT_INT" default:"42" required:"true"`
			TestDefaultBool bool          `env:"TEST_DEFAULT_BOOL" default:"true" required:"true"`
			TestEnvStr      string        `env:"TEST_ENV_STR" required:"true"`
			TestEnvInt      int           `env:"TEST_ENV_INT" required:"true"`
			TestEnvBool     bool          `env:"TEST_ENV_BOOL" required:"true"`
			TestDuration    time.Duration `env:"TEST_DURATION" default:"10s"`
			TestTrueIgnore  int           `ignore:"true"`
			TestFalseIgnore int           `env:"TEST_FALSE_IGNORE"`
		}{},
		envvars: map[string]string{
			"TEST_ENV_STR":      "test env string",
			"TEST_ENV_INT":      "43",
			"TEST_ENV_BOOL":     "1",
			"TEST_FALSE_IGNORE": "1",
		},

		expectedData: map[string]interface{}{
			"TestDefaultStr":  "test string",
			"TestDefaultInt":  42,
			"TestDefaultBool": true,
			"TestEnvStr":      "test env string",
			"TestEnvInt":      43,
			"TestEnvBool":     true,
			"TestDuration":    time.Duration(10 * time.Second),
			"TestTrueIgnore":  0,
			"TestFalseIgnore": 1,
		},
	}, {
		name: "DefaultEnvSource_Err",
		params: &struct {
			TestEnvStr string `env:"TEST_ENV_STR" required:"true"`
		}{},
		envvars: map[string]string{},

		expectedData: map[string]interface{}{
			"TestEnvStr": "",
		},
		expectErr: true,
	}, {
		name: "CustomEnvSource",
		params: &struct {
			TestEnvStr string `env:"TEST_ENV_STR" required:"true"`
		}{},
		envSource: &EnvSource{
			Prefix: "PREFIX_",
		},
		envvars: map[string]string{
			"PREFIX_TEST_ENV_STR": "test env string",
		},

		expectedData: map[string]interface{}{
			"TestEnvStr": "test env string",
		},
	}, {
		name: "CustomEnvSource_Error",
		params: &struct {
			TestEnvStr string `env:"TEST_ENV_STR" required:"true"`
		}{},
		envSource: &EnvSource{
			Prefix: "PREFIX_",
		},
		envvars: map[string]string{
			"TEST_ENV_STR": "test env string",
		},

		expectedData: map[string]interface{}{
			"TestEnvStr": "",
		},
		expectErr: true,
	}, {
		name: "CustomMockSource",
		params: &struct {
			TestStr   string  `mock:"test-str" required:"true"`
			TestInt   int     `mock:"test-int" required:"true"`
			TestBool  bool    `mock:"test-bool" required:"true"`
			TestFloat float64 `mock:"test-float" required:"true"`
		}{},

		mockSource: &mockSource{
			tagKey: "mock",
			vars: map[string]string{
				"test-str":   "test-string",
				"test-int":   "42",
				"test-bool":  "true",
				"test-float": "3.14159",
			},
			expectedLen: 4,
		},

		expectedData: map[string]interface{}{
			"TestStr":   "test-string",
			"TestInt":   42,
			"TestBool":  true,
			"TestFloat": 3.14159,
		},
	}, {
		name: "MissingMockSource",
		params: &struct {
			TestStr   string  `mock:"test-str" required:"true"`
			TestInt   int     `mock:"test-int" required:"true"`
			TestBool  bool    `mock:"test-bool" required:"true"`
			TestFloat float64 `mock:"test-float" required:"true"`
		}{},

		mockSource: nil,

		expectedData: map[string]interface{}{
			"TestStr":   "",
			"TestInt":   0,
			"TestBool":  false,
			"TestFloat": float64(0),
		},
		expectErr: true,
	}, {
		name: "ComplexNestedFields",
		params: &struct {
			TestStr       string   `mock:"test-str" required:"true"`
			TestInt       int      `mock:"test-int" required:"true"`
			TestBool      bool     `mock:"test-bool" required:"true"`
			TestFloat     float64  `mock:"test-float" required:"true"`
			TestStrSlice  []string `mock:"test-strslice" required:"true"`
			TestIntSlice  []int    `mock:"test-intslice" required:"true"`
			TestBoolSlice []bool   `mock:"test-boolslice" required:"true"`

			TestStruct struct {
				NestedStr string `env:"test-str" required:"true"`
			}

			TestInterface interface{}
		}{
			TestInterface: &struct {
				OptionalStr string `env:"TEST_STRING_OPTIONAL"`
			}{},
		},

		envSource: &EnvSource{
			StrictCase: true,
		},
		envvars: map[string]string{
			"test-str": "test-env-string",
		},
		mockSource: &mockSource{
			tagKey: "mock",
			vars: map[string]string{
				"test-str":       "test-string",
				"test-int":       "42",
				"test-bool":      "true",
				"test-float":     "3.14159",
				"test-strslice":  "foo,bar,baz,bat",
				"test-intslice":  "1,2,3,4",
				"test-boolslice": "true,false,0,1",
			},
			expectedLen: 7,
		},

		expectedData: map[string]interface{}{
			"TestStr":       "test-string",
			"TestInt":       42,
			"TestBool":      true,
			"TestFloat":     3.14159,
			"TestStrSlice":  []string{"foo", "bar", "baz", "bat"},
			"TestIntSlice":  []int{1, 2, 3, 4},
			"TestBoolSlice": []bool{true, false, false, true},

			"TestStruct": map[string]interface{}{
				"NestedStr": "test-env-string",
			},
			"TestInterface": map[string]interface{}{
				"OptionalStr": "",
			},
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sources := make([]Source, 0, 2)
			if tc.envSource != nil {
				sources = append(sources, tc.envSource)
			}
			if tc.mockSource != nil {
				sources = append(sources, tc.mockSource)
			}

			os.Clearenv()
			for k, v := range tc.envvars {
				assert.NoError(t, os.Setenv(k, v))
			}

			err := Process(tc.params, sources...)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			data := structs.Map(tc.params)
			assert.Equal(t, tc.expectedData, data)
			tc.mockSource.AssertExpectations(t)
		})
	}
}
