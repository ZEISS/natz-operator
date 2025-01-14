package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeiss/natz-operator/pkg/config"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      config.Config
		expected []byte
	}{
		{
			name:     "empty config",
			cfg:      config.Config{},
			expected: []byte(`{""}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := config.Marshal(tt.cfg)
			require.NoError(t, err)
			fmt.Println(string(b), tt.expected)
			require.Equal(t, tt.expected, b)
		})
	}
}
