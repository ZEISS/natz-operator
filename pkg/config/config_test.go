package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeiss/natz-operator/pkg/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := config.New()
	require.NotNil(t, cfg)
}

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	require.NotNil(t, cfg)

	json, err := cfg.Marshal()
	require.NoError(t, err)
	require.JSONEq(t, `{"host":"0.0.0.0","port":4222}`, string(json))
}
