package hmac

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	// key should be at least 256 bit long, making it
	minimumEntropy = 32

	// the secrets (client and global) should each have at least 16 characters making it harder to guess them
	minimumSecretLength = 32
)

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

// Generate generates a token and a matching signature or returns an error.
// This method implements rfc6819 Section 5.1.4.2.2: Use High Entropy for Secrets.
func (c *HMACToken) Generate() (string, string, error) {
	c.Lock()
	defer c.Unlock()

	if len(c.cfg.SystemSecret) < minimumSecretLength {
		return "", "", errors.Errorf("secret for signing HMAC-SHA256 is expected to be 32 byte long, got %d byte", len(c.cfg.SystemSecret))
	}

	var signingKey [32]byte
	copy(signingKey[:], c.GlobalSecret)

	if c.cfg.TokenEntropy < minimumEntropy {
		c.cfg.TokenEntropy = minimumEntropy
	}

	tokenKey, err := RandomBytes(c.cfg.TokenEntropy)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	signature := generateHMAC(tokenKey, &signingKey)

	encodedSignature := b64.EncodeToString(signature)
	encodedToken := fmt.Sprintf("%s.%s", b64.EncodeToString(tokenKey), encodedSignature)
	return encodedToken, encodedSignature, nil
}

// Validate validates a token and returns its signature or an error if the token is not valid.
func (c *HMACToken) Validate(token string) (err error) {
	if err = c.validate(c.GlobalSecret, token); err == nil {
		return nil
	} else if errors.Is(err, ErrTokenSignatureMismatch) {
	} else {
		return err
	}

	if err == nil {
		return errors.New("a secret for signing HMAC-SHA256 is expected to be defined, but none were")
	}

	return err
}

func (c *HMACToken) validate(secret []byte, token string) error {
	fmt.Println(token)
	if len(secret) < minimumSecretLength {
		return errors.Errorf("secret for signing HMAC-SHA256 is expected to be 32 byte long, got %d byte", len(secret))
	}

	var signingKey [32]byte
	copy(signingKey[:], secret)

	split := strings.Split(token, ".")
	if len(split) != 2 {
		return errors.WithStack(ErrInvalidTokenFormat)
	}

	tokenKey := split[0]
	tokenSignature := split[1]
	if tokenKey == "" || tokenSignature == "" {
		return errors.WithStack(ErrInvalidTokenFormat)
	}

	decodedTokenSignature, err := b64.DecodeString(tokenSignature)
	if err != nil {
		return errors.WithStack(err)
	}

	decodedTokenKey, err := b64.DecodeString(tokenKey)
	if err != nil {
		return errors.WithStack(err)
	}

	expectedMAC := generateHMAC(decodedTokenKey, &signingKey)
	if !hmac.Equal(expectedMAC, decodedTokenSignature) {
		// Hash is invalid
		return errors.WithStack(ErrTokenSignatureMismatch)
	}

	return nil
}

func (c *HMACToken) Signature(token string) string {
	split := strings.Split(token, ".")

	if len(split) != 2 {
		return ""
	}

	return split[1]
}

func generateHMAC(data []byte, key *[32]byte) []byte {
	h := hmac.New(sha512.New512_256, key[:])
	// sha512.digest.Write() always returns nil for err, the panic should never happen
	_, err := h.Write(data)
	if err != nil {
		panic(err)
	}
	return h.Sum(nil)
}
