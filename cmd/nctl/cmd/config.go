package cmd

import (
	"os"
	"path/filepath"

	"github.com/zeiss/pkg/utilx"
	"k8s.io/client-go/util/homedir"
)

// DefaultConfig is the default configuration for the root command.
func DefaultConfig() Config {
	return Config{
		Verbose:   false,
		Force:     false,
		Namespace: "default",
	}
}

// Config is a struct that holds the configuration for the root command.
type Config struct {
	KubeConfig string           `json:"kubeconfig"`
	Context    string           `json:"context"`
	Verbose    bool             `json:"verbose"`
	Force      bool             `json:"force"`
	Creds      CredsConfig      `json:"creds"`
	Activation ActivationConfig `json:"activation"`
	Namespace  string           `json:"namespace"`
}

// GetKubeConfig returns the path to the kubeconfig file.
func (c *Config) GetKubeConfig() string {
	return utilx.IfElse(
		utilx.NotEmpty(homedir.HomeDir()),
		filepath.Join(homedir.HomeDir(), ".kube", "config"),
		c.KubeConfig,
	)
}

// Cwd is the current working directory.
func (c *Config) Cwd() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return p, nil
}
