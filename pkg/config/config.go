package config

// New returns a new Config object.
func New() *Config {
	return &Config{}
}

// Config ...
type Config struct {
	Properties []*Property
}

// Property ...
type Property struct {
	// Name ...
	Name string `json:"name"`
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

// Block_Object ...
type Block_Object struct{}

// Block_Array ...
type Block_Array struct{}

// Block_Include ...
type Block_Include struct{}

// Block_String ...
type Block_String struct{}

func (b *Block_Object) isBlock_Block() {}

func (b *Block_Array) isBlock_Block() {}

func (b *Block_String) isBlock_Block() {}

func (b *Block_Include) isBlock_Block() {}
