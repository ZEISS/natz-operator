package config

// New returns a new Config object.
func New() *Config {
	return &Config{}
}

// Config ...
type Config struct {
	// Host ...
	Host string `json:"host,omitempty"`
	// Port ...
	Port int `json:"port,omitempty"`
	// Gateway ...
	Gateway Gateway `json:"gateway,omitempty"`
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

// Marhshal ...
func (c *Config) Marshal() ([]byte, error) {
	return nil, nil
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
