package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeiss/natz-operator/pkg/config"
)

func TestNewWriter(t *testing.T) {
	t.Parallel()

	cfg := config.New()
	require.NotNil(t, cfg)
}

func TestConfiguration(t *testing.T) {
	t.Parallel()

	cfg := &config.Property{
		Block: &config.Block_Object{},
	}
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.GetBlock())
	require.Equal(t, cfg.Block, cfg.GetBlock())
	require.Equal(t, &config.Block_Object{}, cfg.GetBlock())
}
