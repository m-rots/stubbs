// Package stubbs manages the authentication process of Google Service Accounts.
package stubbs

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"sync"
)

// Stubbs manages the authentication process of Google Service Accounts.
type Stubbs struct {
	mtx sync.Mutex // Mutex to make sure one refresh happens at a time.

	clientEmail string          // Service account client email address.
	privateKey  *rsa.PrivateKey // Service account RSA private key.
	scopes      []string        // Authentication scopes.

	exp   int64  // UNIX expiration time of the cached token.
	token string // Cached access token.

	lifetime    int64 // Lifetime of the access token in seconds.
	safeRefresh int64
}

// An Option can adjust default Stubbs values.
type Option func(*Stubbs)

// WithLifetime overwrites the default lifetime (3600) and safeRefresh (10) values.
//
// Lifetime is the lifetime of the access token in seconds.
// It MUST be set between 1 and 3600.
//
// SafeRefresh indicates how far ahead in seconds the cache should be invalidated.
// The safeRefresh value MUST be lower than the lifetime.
func WithLifetime(lifetime, safeRefresh int64) Option {
	return func(s *Stubbs) {
		s.lifetime = lifetime
		s.safeRefresh = safeRefresh
	}
}

// New creates a new instance of Stubbs.
//
// A new Stubbs should be created for different service accounts,
// service account keys and authentication scopes.
//
// The same Stubbs can be reused when a token expires,
// as the access token is automatically refreshed when the present time
// surpasses the expiration time of the cached access token.
func New(email string, priv *rsa.PrivateKey, scopes []string, opts ...Option) *Stubbs {
	s := &Stubbs{
		clientEmail: email,
		privateKey:  priv,
		scopes:      scopes,
		lifetime:    3600,
		safeRefresh: 10,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ParseKey is a utility function to parse a string into a RSA PrivateKey.
func ParseKey(priv string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(priv))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch key := key.(type) {
	case *rsa.PrivateKey:
		return key, nil
	default:
		return nil, errors.New("Invalid key type")
	}
}

// FromFile creates a new instance of Stubbs
// from a Google Service Account JSON key.
func FromFile(filename string, scopes []string, opts ...Option) (*Stubbs, error) {
	type serviceAccount struct {
		Email string `json:"client_email"`
		Key   string `json:"private_key"`
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	sa := new(serviceAccount)
	err = json.NewDecoder(file).Decode(sa)
	if err != nil {
		return nil, err
	}

	priv, err := ParseKey(sa.Key)
	if err != nil {
		return nil, err
	}

	return New(sa.Email, priv, scopes, opts...), nil
}
