package v1alpha1

import (
	"github.com/zeiss/pkg/utilx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigPhase string

const (
	// SecretConfigDataKey is the key for the config in the secret
	SecretConfigDataKey = "nats.conf"
)

// New returns a new Config object.
func New() *Config {
	return &Config{
		Resolver: Resolver{},
	}
}

// Default returns the default configuration.
func Default() *Config {
	cfg := New()

	utilx.SetDefaults(cfg)

	return cfg
}

// Config ...
type Config struct {
	// Host ...
	Host string `json:"host,omitempty" default:"0.0.0.0"`
	// Port ...
	Port int `json:"port,omitempty" default:"4222"`
	// HTTPPort ...
	HTTPPort int `json:"http_port,omitempty" default:"8222"`
	// Gateway ...
	Gateway *Gateway `json:"gateway,omitempty"`
	// ClientAdvertise ...
	ClientAdvertise string `json:"client_advertise,omitempty"`
	// TLS ...
	TLS *TLS `json:"tls,omitempty"`
	// Authorization ...
	Authorization *Authorization `json:"authorization,omitempty"`
	// Resolver ...
	Resolver Resolver `json:"resolver,omitempty"`
	// ResolverPreload ...
	ResolverPreload ResolverPreload `json:"resolver_preload,omitempty"`
	// SystemAccount ...
	SystemAccount string `json:"system_account,omitempty"`
	// Operator ...
	Operator string `json:"operator,omitempty"`
	// PidFile ...
	PidFile string `json:"pid_file,omitempty" default:"/var/run/nats/nats.pid"`
}

// Resolver ...
type Resolver struct {
	// Type ...
	Type string `json:"type,omitempty" default:"full"`
	// Dir ...
	Dir string `json:"dir,omitempty" default:"/data/resolver"`
	// AllowedDelete ...
	AllowedDelete bool `json:"allowed_delete,omitempty" default:"true"`
	// Interval ...
	Interval string `json:"interval,omitempty" default:"2m"`
	// Limit ...
	Limit int `json:"limit,omitzero"`
	// Timeout ...
	Timeout string `json:"timeout,omitempty" default:"5s"`
}

// ResolverPreload ...
type ResolverPreload map[string]string

// JetStream ...
type JetStream struct {
	// Enabled ...
	Enabled bool `json:"enabled" default:"true"`
	// StoreDir ...
	StoreDir string `json:"store_dir" default:"/tmp/nats/jetstream"`
	// MaxMemoryStore ...
	MaxMemoryStore int `json:"max_memory_store,omitempty"`
	// MaxFileStore ...
	MaxFileStore int `json:"max_file_store,omitempty"`
	// Domain ...
	Domain string `json:"domain,omitempty"`
	// EncryptionKey ...
	EncryptionKey string `json:"encryption_key,omitempty"`
	// Cipher ...
	Cipher string `json:"cipher,omitempty"`
	// ExtensionHint ...
	ExtensionHint string `json:"extension_hint,omitempty"`
	// Limits ...
	Limits JetStreamLimits `json:"limits,omitempty"`
	// UniqueTag ...
	UniqueTag string `json:"unique_tag,omitempty"`
	// MaxOutStandingCatchUp ...
	MaxOutStandingCatchUp string `json:"max_outstanding_catchup,omitempty" default:"32M"`
	// SyncInterval ...
	SyncInterval string `json:"sync_interval,omitempty" default:"2m"`
}

// JetStreamLimits ...
type JetStreamLimits struct {
	// MaxAckPending ...
	MaxAckPending int `json:"max_ack_pending,omitempty"`
	// MaxHaAssets ...
	MaxHaAssets int `json:"max_ha_assets,omitempty"`
	// MaxRequestBatch ...
	MaxRequestBatch int `json:"max_request_batch,omitempty"`
	// DuplicateWindow ...
	DuplicateWindow int `json:"duplicate_window,omitempty"`
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
	CipherSuites string `json:"cipher_suites"`
	// CurvePreferences ...
	CurvePreferences string `json:"curve_preferences,omitempty"`
	// Insecure ...
	Insecure bool `json:"insecure,omitempty"`
	// Verify ...
	Verify bool `json:"verify"`
	// VerifyAndMap ...
	VerifyAndMap bool `json:"verify_and_map"`
	// VerifyCertAndCheckKnownURLs ...
	VerifyCertAndCheckKnownURLs bool `json:"verify_cert_and_check_known_urls,omitempty"`
	// ConnectionRateLimit ...
	ConnectionRateLimit int `json:"connection_rate_limit,omitempty"`
	// PinnedCerts ...
	PinnedCerts []string `json:"pinned_certs"`
}

// Gateway ...
type Gateway struct {
	// Name ...
	Name string `json:"name"`
	// RejectUnknownCluster ...
	RejectUnknownCluster bool `json:"reject_unknown_cluster,omitempty"`
	// Authorization ...
	Authorization Authorization `json:"authorization,omitempty"`
	// Host ...
	Host string `json:"host,omitempty"`
	// Port ...
	Port int `json:"port,omitempty"`
	// Listen ...
	Listen string `json:"listen,omitempty"`
	// Advertise ...
	Advertise string `json:"advertise,omitempty"`
	// ConnectTimeout ...
	ConnectRetries int `json:"connect_retries,omitempty"`
	// Gateways ...
	Gateways []GatewayEntry `json:"gateways,omitempty"`
}

// GatewayEntry ...
type GatewayEntry struct {
	// Name ...
	Name string `json:"name"`
	// URLS ...
	URLS []string `json:"urls"`
	// TLS ...
	TLS TLS `json:"tls,omitempty"`
}

// Authorization ...
type Authorization struct {
	User        string      `json:"user,omitempty"`
	Password    string      `json:"password,omitempty"`
	Token       string      `json:"token,omitempty"`
	Timeout     int         `json:"timeout,omitempty"`
	AuthCallout AuthCallout `json:"auth_callout,omitempty"`
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

const (
	ConfigPhaseNone         ConfigPhase = ""
	ConfigPhasePending      ConfigPhase = "Pending"
	ConfigPhaseCreating     ConfigPhase = "Creating"
	ConfigPhaseSynchronized ConfigPhase = "Synchronized"
	ConfigPhaseFailed       ConfigPhase = "Failed"
)

// NatsConfigSpec defines the desired state of NatsConfig
type NatsConfigSpec struct {
	// OperatorRef is a reference to the operator that is managing the config.
	OperatorRef NatsOperatorReference `json:"operatorRef"`
	// SystemAccountRef is a reference to the system account.
	SystemAccountRef NatsAccountReference `json:"systemAccountRef"`
	// Gateways is a list of gateways that should be configured.
	Gateways []NatsgatewayReference `json:"gateways,omitempty"`
	// Config is the configuration that should be applied.
	Config Config `json:"config,omitempty"`
}

// NatsConfigStatus defines the observed state of NatsConfig
type NatsConfigStatus struct {
	// Conditions is an array of conditions that the operator is currently in.
	Conditions []metav1.Condition `json:"conditions,omitempty" optional:"true"`
	// Phase is the current phase of the operator.
	//
	// +kubebuilder:validation:Enum={None,Pending,Creating,Synchronized,Failed}
	Phase ConfigPhase `json:"phase"`
	// ControlPaused is a flag that indicates if the operator is paused.
	ControlPaused bool `json:"controlPaused,omitempty" optional:"true"`
	// LastUpdate is the timestamp of the last update.
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +genclient
// +genclient:nonNamespaced
// +genreconciler

type NatsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatsConfigSpec   `json:"spec,omitempty"`
	Status NatsConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NatsConfigList contains a list of NatsConfig
type NatsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NatsConfig{}, &NatsConfigList{})
}
