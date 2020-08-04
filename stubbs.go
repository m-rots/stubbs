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
	clientEmail string          // service account client email address
	lifetime    int             // token lifetime in seconds
	privateKey  *rsa.PrivateKey // service account RSA private key
	scopes      []string        // authentication scopes

	exp   int64  // the UNIX expiration time of the cached token
	token string // the cached access token

	mtx sync.Mutex // mutex to make sure one refresh happens at a time
}

// New creates a new instance of Stubbs.
//
// A new Stubbs should be created for different service accounts,
// service account keys, expiration values and authentication scopes.
//
// The same Stubbs can be reused when a token expires,
// as the access token is automatically refreshed when the present time
// surpasses the expiration time of the cached access token.
func New(email string, priv *rsa.PrivateKey, scopes []string, lifetime int) *Stubbs {
	return &Stubbs{
		clientEmail: email,
		privateKey:  priv,
		scopes:      scopes,
		lifetime:    lifetime,
	}
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
func FromFile(filename string, scopes []string, lifetime int) (*Stubbs, error) {
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

	return New(sa.Email, priv, scopes, lifetime), nil
}
