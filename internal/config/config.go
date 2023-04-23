package config

// Config represents generic service configuration.
type Config interface {
	UnmarshalJSON(data []byte) error
	String() string
}
