package config

type Config struct {
	Value1 string
	Value2 string
}

func (c *Config) GetValue1() string {
	return c.Value1
}

func (c *Config) SetValue1(value1 string) {
	c.Value1 = value1
}

func (c *Config) GetValue2() string {
	return c.Value2
}

func (c *Config) SetValue2(value2 string) {
	c.Value2 = value2
}

func NewConfig() *Config {
	return new(Config)
}

func NewConfigAsInterface() interface{} {
	return NewConfig()
}
