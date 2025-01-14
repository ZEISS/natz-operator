package config_test

import (
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
			expected: []byte(`{host:""}`),
		},
		{
			name: "config with Host",
			cfg: config.Config{
				Host: "nats://localhost:4222",
			},
			expected: []byte(`{host:"nats://localhost:4222"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := config.Marshal(tt.cfg)
			require.NoError(t, err)
			require.Equal(t, tt.expected, b)
		})
	}
}
