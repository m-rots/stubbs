package stubbs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// AccessToken returns a new or cached (but not expired) access token
// and the token's expiry time in UNIX.
//
// It may produce an error when a bad request is made to Google's OAuth server.
// Such a bad request can occur when invalid scopes are given,
// the key of the service account is deleted
// or when the service account itself is deleted.
func (acc *Stubbs) AccessToken() (string, int64, error) {
	if acc.token == "" || time.Now().Unix() > acc.exp {
		err := acc.refreshToken()
		if err != nil {
			return "", 0, err
		}
	}

	return acc.token, acc.exp, nil
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
}

// refreshToken refreshes the access token which is internally stored
// inside the Stubbs instance and updates `exp` and `token`.
func (acc *Stubbs) refreshToken() error {
	jwt, exp := acc.createJWT()

	res, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	})

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Received status code: %v", res.StatusCode)
	}

	response := new(oauthResponse)

	decoder := json.NewDecoder(res.Body)
	if decoder.Decode(response) != nil {
		return fmt.Errorf("JSON decoding error")
	}

	if response.AccessToken == "" {
		return fmt.Errorf("Did not retrieve access token")
	}

	acc.exp = exp
	acc.token = response.AccessToken

	return nil
}
