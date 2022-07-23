package ssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/onetwentyseven-dev/go-config"
	"github.com/stretchr/testify/assert"
)

type mockParameter struct {
	val string
	err error

	expectVal bool
}

func (m *mockParameter) NoValue() error {
	return nil
}

func (m *mockParameter) SetValue(val string) error {
	m.val = val
	return m.err
}

func (m *mockParameter) AssertExpectations(t *testing.T) {
	if m.expectVal {
		assert.NotEmpty(t, m.val)
	} else {
		assert.Empty(t, m.val)
	}
}

type mockSsm struct {
	params map[string]string
	err    error
}

func (m *mockSsm) GetParameters(_ context.Context, in *ssm.GetParametersInput, _ ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	out := ssm.GetParametersOutput{
		Parameters: make([]types.Parameter, 0, len(in.Names)),
	}
	for _, n := range in.Names {
		if p, ok := m.params[n]; ok {
			out.Parameters = append(out.Parameters, types.Parameter{
				Name:  aws.String(n),
				Value: aws.String(p),
			})
		}
	}

	return &out, nil
}

func TestSource_TagKey(t *testing.T) {
	var src Source
	assert.Equal(t, "ssm", src.TagKey())
}

func TestSource_Process(t *testing.T) {
	testCases := []struct {
		name string

		params map[string][]*mockParameter
		mock   mockSsm
		prefix string

		expectErr bool
	}{{
		name: "ErrGetParameters",
		params: map[string][]*mockParameter{
			"key": {{
				expectVal: false,
			}},
		},
		mock: mockSsm{
			err: errors.New("test error"),
		},
		expectErr: true,
	}, {
		name: "ErrSetParameter",
		params: map[string][]*mockParameter{
			"KEY": {{
				err:       errors.New("test error"),
				expectVal: true,
			}},
		},
		prefix: "/test/prefix/",
		mock: mockSsm{
			params: map[string]string{
				"/test/prefix/key": "value",
			},
		},
		expectErr: true,
	}, {
		name: "NormalGet",
		params: map[string][]*mockParameter{
			"KEY": {{
				expectVal: true,
			}},
			"/test/prefix/key2": {{
				expectVal: false,
			}},
			"/different-prefix/key3,absolute": {{
				expectVal: true,
			}},
		},
		prefix: "/test/prefix/",
		mock: mockSsm{
			params: map[string]string{
				"/test/prefix/key":       "value",
				"/different-prefix/key3": "value-2",
			},
		},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paramMap := make(map[string][]config.Parameter, len(tc.params))
			for k, ps := range tc.params {
				params := make([]config.Parameter, len(ps))
				for i, p := range ps {
					params[i] = p
				}

				paramMap[k] = params
			}

			source := New(tc.prefix, &tc.mock)
			err := source.Process(paramMap)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, ps := range tc.params {
				for _, p := range ps {
					p.AssertExpectations(t)
				}
			}
		})
	}
}
