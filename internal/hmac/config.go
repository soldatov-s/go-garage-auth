package hmac

import "time"

type Config struct {
	TokenEntropy int           `envconfig:"default=100"`
	SystemSecret string        `envconfig:"default=you_Really_Need_To_ChangeThis!!!"`
	TTL          time.Duration `envconfig:"default=24h"`
}
