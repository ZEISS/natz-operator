package v1alpha1_test

import (
	"testing"

	"github.com/zeiss/natz-operator/api/v1alpha1"

	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	got := v1alpha1.Default()
	want := &v1alpha1.Config{
		Host:     "0.0.0.0",
		Port:     4222,
		HTTPPort: 8222,
	}

	require.Equal(t, want, got)
}
