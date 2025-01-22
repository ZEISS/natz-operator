package config

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/zeiss/pkg/cast"
)

// New returns a new Config object.
func New() *Config {
	return &Config{}
}

// Default ...
func Default() *Config {
	return &Config{
		Host: cast.Ptr("0.0.0.0"),
		Port: cast.Ptr(4222),
	}
}

// Config ...
type Config struct {
	// Host ...
	Host *string `json:"host,omitempty"`
	// Port ...
	Port *int `json:"port,omitempty"`
	// HTTPPort ...
	HTTPPort *int `json:"http_port,omitempty"`
	// Gateway ...
	Gateway *Gateway `json:"gateway,omitempty"`
	// ClientAdvertise ...
	ClientAdvertise *string `json:"client_advertise,omitempty"`
	// TLS ...
	TLS *TLS `json:"tls,omitempty"`
}

// JetStream ...
type JetStream struct {
	// Enabled ...
	Enabled bool `json:"enabled" default:"true"`
	// StoreDir ...
	StoreDir string `json:"store_dir" default:"/tmp/nats/jetstream"`
	// MaxMemoryStore ...
	MaxMemoryStore *int `json:"max_memory_store,omitempty"`
	// MaxFileStore ...
	MaxFileStore *int `json:"max_file_store,omitempty"`
	// Domain ...
	Domain *string `json:"domain,omitempty"`
	// EncryptionKey ...
	EncryptionKey *string `json:"encryption_key,omitempty"`
	// Cipher ...
	Cipher *string `json:"cipher,omitempty"`
	// ExtensionHint ...
	ExtensionHint *string `json:"extension_hint,omitempty"`
	// Limits ...
	Limits *JetStreamLimits `json:"limits,omitempty"`
	// UniqueTag ...
	UniqueTag *string `json:"unique_tag,omitempty"`
	// MaxOutStandingCatchUp ...
	MaxOutStandingCatchUp *string `json:"max_outstanding_catchup,omitempty" default:"32M"`
	// SyncInterval ...
	SyncInterval *string `json:"sync_interval,omitempty" default:"2m"`
}

// JetStreamLimits ...
type JetStreamLimits struct {
	// MaxAckPending ...
	MaxAckPending *int `json:"max_ack_pending,omitempty"`
	// MaxHaAssets ...
	MaxHaAssets *int `json:"max_ha_assets,omitempty"`
	// MaxRequestBatch ...
	MaxRequestBatch *int `json:"max_request_batch,omitempty"`
	// DuplicateWindow ...
	DuplicateWindow *int `json:"duplicate_window,omitempty"`
}

// TLS ...
type TLS struct {
	// CertFile ...
	CertFile string `json:"cert_file"`
	// KeyFile ...
	KeyFile string `json:"key_file"`
	// CAFile ...
	CAFile string `json:"ca_file"`
	// CipherSuites ...
	CipherSuites *string `json:"cipher_suites"`
	// CurvePreferences ...
	CurvePreferences *string `json:"curve_preferences,omitempty"`
	// Insecure ...
	Insecure *bool `json:"insecure,omitempty"`
	// Verify ...
	Verify bool `json:"verify"`
	// VerifyAndMap ...
	VerifyAndMap bool `json:"verify_and_map"`
	// VerifyCertAndCheckKnownURLs ...
	VerifyCertAndCheckKnownURLs *bool `json:"verify_cert_and_check_known_urls,omitempty"`
	// ConnectionRateLimit ...
	ConnectionRateLimit *int `json:"connection_rate_limit,omitempty"`
	// PinnedCerts ...
	PinnedCerts []string `json:"pinned_certs"`
}

// Marshal ...
func (c *Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// Unmarshal ...
func (c *Config) Unmarshal(data []byte) error {
	cfg := struct {
		Host            *string `json:"host,omitempty"`
		Port            *int    `json:"port,omitempty"`
		HTTPPort        *int    `json:"http_port,omitempty"`
		Gateway         *Gateway
		ClientAdvertise *string `json:"client_advertise,omitempty"`
		TLS             *TLS    `json:"tls,omitempty"`
	}{}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return errors.WithStack(err)
	}

	c.Host = cfg.Host
	c.Port = cfg.Port
	c.HTTPPort = cfg.HTTPPort
	c.Gateway = cfg.Gateway
	c.ClientAdvertise = cfg.ClientAdvertise
	c.TLS = cfg.TLS

	return nil
}

// Gateway ...
type Gateway struct {
	// Name ...
	Name string `json:"name"`
	// RejectUnknownCluster ...
	RejectUnknownCluster *bool `json:"reject_unknown_cluster,omitempty"`
	// Authorization ...
	Authorization *Authorization `json:"authorization,omitempty"`
	// Host ...
	Host *string `json:"host,omitempty"`
	// Port ...
	Port *int `json:"port,omitempty"`
	// Listen ...
	Listen *string `json:"listen,omitempty"`
	// Advertise ...
	Advertise *string `json:"advertise,omitempty"`
	// ConnectTimeout ...
	ConnectRetries *int `json:"connect_retries,omitempty"`
	// Gateways ...
}

// GatewayEntry ...
type GatewayEntry struct {
	// Name ...
	Name string `json:"name"`
	// URLS ...
	URLS []string `json:"urls"`
	// TLS ...
	TLS *TLS `json:"tls,omitempty"`
}

// Authorization ...
type Authorization struct {
	User        *string      `json:"user,omitempty"`
	Password    *string      `json:"password,omitempty"`
	Token       *string      `json:"token,omitempty"`
	Timeout     *int         `json:"timeout,omitempty"`
	AuthCallout *AuthCallout `json:"auth_callout,omitempty"`
}

// AuthCallout ...
type AuthCallout struct {
	// Issuer ...
	Issuer string `json:"issuer"`
	// AuthUsers ...
	AuthUsers []string `json:"auth_users"`
	// Account ...
	Account string `json:"account"`
	// XKey ...
	XKey string `json:"xkey"`
}

// Property ...
type Property struct {
	// Name ...
	Name string
	// Block is the configuration block.
	Block isBlock_Block
}

// Block is an interface for a configuration block.
type Block interface {
	isBlock_Block()
}

// GetBlock ...
func (c *Property) GetBlock() isBlock_Block {
	if c != nil {
		return c.Block
	}

	return nil
}

type isBlock_Block interface{}

// Block_Object represents an object of a configuration block.
type Block_Object struct{}

// Block_Array represents an array of a configuration block.
type Block_Array struct{}

// Block_Include
type Block_Include struct{}

// Block_String ...
type Block_String struct {
	// Value ...
	Value string
}

func (b *Block_Object) isBlock_Block() {}

func (b *Block_Array) isBlock_Block() {}

func (b *Block_String) isBlock_Block() {}

func (b *Block_Include) isBlock_Block() {}
