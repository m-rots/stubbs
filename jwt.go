package stubbs

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// The header of a JSON Web Token
type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// The payload (claims) of a JSON Web Token.
//
// The claims of this payload are specific to Google's OAuth
type payload struct {
	Iss   string `json:"iss"`
	Scope string `json:"scope"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

// toBase64 converts any interface into its JSON representation
// encoded in Base64.
func toBase64(content interface{}) string {
	contentJSON, _ := json.Marshal(content)
	return base64.RawURLEncoding.EncodeToString(contentJSON)
}

// createSignature creates a signature for any message with the given RSA private key
// using the RSA SHA-256 algorithm.
func createSignature(message string, priv *rsa.PrivateKey) string {
	hashed := sha256.Sum256([]byte(message))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	return base64.RawURLEncoding.EncodeToString(signature)
}

// createJWT creates a RS256 Google OAuth specific JSON Web Token (JWT)
// using the email and RSA private key of a service account.
// It returns both the Base64 encoded JWT as well as the expiry time in UNIX.
func (acc *Stubbs) createJWT() (string, int64) {
	now := time.Now().Unix()
	expires := now + int64(acc.lifetime)

	header := toBase64(header{
		Alg: "RS256",
		Typ: "JWT",
	})

	payload := toBase64(payload{
		Aud:   "https://oauth2.googleapis.com/token",
		Iss:   acc.clientEmail,
		Scope: strings.Join(acc.scopes, " "),
		Iat:   now,
		Exp:   expires,
	})

	message := header + "." + payload
	signature := createSignature(message, acc.privateKey)
	jwt := message + "." + signature

	return jwt, expires
}
