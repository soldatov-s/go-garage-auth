package hmac

import (
	"context"
	"crypto/sha256"
	"sync"

	"github.com/soldatov-s/go-garage/domains"
)

const (
	DomainName = "hmac"
)

type HMACToken struct {
	cfg          *Config
	GlobalSecret []byte
	sync.Mutex
}

// HashStringSecret hashes the secret for consumption by the AEAD encryption algorithm which expects exactly 32 bytes.
//
// The system secret is being hashed to always match exactly the 32 bytes required by AEAD, even if the secret is long or
// shorter.
func HashStringSecret(secret string) []byte {
	return HashByteSecret([]byte(secret))
}

// HashByteSecret hashes the secret for consumption by the AEAD encryption algorithm which expects exactly 32 bytes.
//
// The system secret is being hashed to always match exactly the 32 bytes required by AEAD, even if the secret is long or
// shorter.
func HashByteSecret(secret []byte) []byte {
	r := sha256.Sum256(secret)
	return r[:]
}

func Registrate(ctx context.Context, cfg *Config) (context.Context, error) {
	t := &HMACToken{
		cfg:          cfg,
		GlobalSecret: HashStringSecret(cfg.SystemSecret),
	}

	return domains.RegistrateByName(ctx, DomainName, t), nil
}

func Get(ctx context.Context) (*HMACToken, error) {
	if v, ok := domains.GetByName(ctx, DomainName).(*HMACToken); ok {
		return v, nil
	}
	return nil, domains.ErrInvalidDomainType
}
