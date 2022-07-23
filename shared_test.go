package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockParameter struct {
	noValCalled bool
	noValErr    error

	setVal    string
	setValErr error

	expectValue bool
}

func (m *mockParameter) NoValue() error {
	m.noValCalled = true
	return m.noValErr
}

func (m *mockParameter) SetValue(v string) error {
	if v == "" {
		return m.NoValue()
	}

	m.setVal = v
	return m.setValErr
}

func (m *mockParameter) AssertExpectations(t *testing.T) {
	if m.expectValue {
		assert.False(t, m.noValCalled)
		assert.NotEmpty(t, m.setVal)
	} else {
		assert.True(t, m.noValCalled)
		assert.Empty(t, m.setVal)
	}
}

type mockSource struct {
	tagKey string
	vars   map[string]string

	processInput map[string][]Parameter

	expectedLen int
}

func (m *mockSource) TagKey() string {
	return m.tagKey
}

func (m *mockSource) Process(input map[string][]Parameter) error {
	m.processInput = input

	for k, params := range input {
		v := m.vars[k]
		for _, p := range params {
			if err := p.SetValue(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *mockSource) AssertExpectations(t *testing.T) {
	if m == nil {
		return
	}

	assert.Len(t, m.processInput, m.expectedLen)
}
