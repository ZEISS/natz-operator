package v1alpha1_test

import (
	"testing"

	"github.com/zeiss/natz-operator/api/v1alpha1"

	"github.com/stretchr/testify/require"
	"github.com/zeiss/pkg/copy"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	got := v1alpha1.Default()
	want := &v1alpha1.Config{
		Host:     "0.0.0.0",
		Port:     4222,
		HTTPPort: 8222,
		PidFile:  "/var/run/nats/nats.pid",
		Resolver: v1alpha1.Resolver{
			Type:          "full",
			Dir:           "/data/resolver",
			AllowedDelete: true,
			Interval:      "2m",
			Timeout:       "5s",
		},
	}

	require.Equal(t, want, got)

	other := &v1alpha1.Config{
		HTTPPort: 8223,
	}

	want = &v1alpha1.Config{
		Host:     "0.0.0.0",
		Port:     4222,
		HTTPPort: 8223,
		PidFile:  "/var/run/nats/nats.pid",
		Resolver: v1alpha1.Resolver{
			Type:          "full",
			Dir:           "/data/resolver",
			AllowedDelete: true,
			Interval:      "2m",
			Timeout:       "5s",
		},
	}

	err := copy.CopyWithOption(got, other, copy.WithIgnoreEmpty())
	require.NoError(t, err)
	require.Equal(t, want, got)
}
