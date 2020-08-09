<p align="center"><img width=450 src="banner.svg" /></p>

## Introduction

Stubbs plays a key role in my Journey of Transfer narrative as the Head of Security.
Specifically, he manages the authentication process of Google's [Service Accounts](https://cloud.google.com/iam/docs/service-accounts#what_are_service_accounts).

Journey of Transfer is a narrative I am writing with projects named after characters of Westworld. The narrative is my exploration process of the [Go language](https://golang.org), while building a programme utilising service accounts to upload and sync files to Google Drive.

Stubbs is the first character of this narrative and was created because the authentication process is essential to all other branches of the narrative. I grew tired of having to rewrite the authentication module over and over again in my personal projects, of which a majority use Google's APIs in some way. To speed up my development workflow, I have included a small command line interface to create new access tokens straight from the terminal.

## Using the CLI

The command line interface (CLI) can create a one-off access token for any service account. The CLI requires a JSON key file of a service account to automatically read the client email and private key fields. To supply this key file you need to supply the path of the file as the first and only argument.

Additionally, you need to supply at least one authentication scope to the CLI with the `--scope` or `-s` flag. As most, if not all, authentication scopes start with `https://www.googleapis.com/auth/`, you can skip the URL part and only supply the scope itself such as `drive.readonly`. Full URLs still work so the whole scope can be provided as well. To supply multiple scopes you can use the `--scope` flag multiple times.

Moreover, the access token has a maximum lifetime of 3600 seconds (1 hour). By default the maximum lifetime is used, though a custom lifetime, defined in seconds, can be supplied with the `--lifetime` or `-l` flag.

Finally, the CLI renders a small overview of the parsed email address, authentication scopes, the resulting access token and a countdown of the remaining lifetime of the access token.

### Installing the CLI

If you already have Go installed you can run the following to install the CLI globally:

```bash
GO111MODULE=off go get -u github.com/m-rots/stubbs/cmd/stubbs
```

### CLI example

```bash
stubbs -s "iam" -s "drive.readonly" -s "https://mail.google.com/" --lifetime 60 sa.json
```

In this example the lifetime of the access token is 60 seconds, the JSON key file of the service account is located at `sa.json` and the following three scopes are used:

1. `https://www.googleapis.com/auth/iam`
2. `https://www.googleapis.com/auth/drive.readonly`
3. `https://mail.google.com/`

## Using the module

**Important: In contrast to the CLI, the module requires the full authentication scope URLs and caches the access token by default.**

Stubbs, as a module, only has three core drives:

1. Create a JSON Web Token with a Google OAuth specific claim set
2. Make a request to the Google OAuth endpoint to fetch an access token
3. Cache the access token response and refresh the cache once the token expires

To initialise an instance of Stubbs, you have to supply the client email and RSA key of a service account, a list of authentication scopes and the desired lifetime of the access token in seconds. If you are reading the RSA key from a PEM encoded string, you can use the `ParseKey()` function to convert the string into a `rsa.PrivateKey`, which is a valid input for the initialisation of Stubbs.

A new instance of Stubbs should be created when using a different service account, service account key or set of scopes. Usually you only reuse an instance of Stubbs when the lifetime of your programme outruns the lifetime of the access token, as the `AccessToken()` method automatically refreshes the access token when it has surpassed its lifetime.

### Installing the module

To add Stubbs to your project, simply run:

```bash
go get github.com/m-rots/stubbs
```

### Module example

The example does not work by default. Please replace the clientEmail and privateKey variables with valid service account counterparts.

```golang
package main

import (
  "fmt"
  "os"
  
  "github.com/m-rots/stubbs"
)

func main() {
  clientEmail := "stubbs@westworld.iam.gserviceaccount.com"
  privateKey := "-----BEGIN PRIVATE KEY-----\n..."
  scopes := []string{
    "https://www.googleapis.com/auth/iam",
    "https://www.googleapis.com/auth/drive.readonly",
    "https://mail.google.com/",
  }

  priv, err := stubbs.ParseKey(privateKey)
  if err != nil {
    fmt.Println("Invalid private key")
    os.Exit(1)
  }

  account := stubbs.New(clientEmail, &priv, scopes)

  token, exp, err := account.AccessToken()
  if err != nil {
    fmt.Println("Error retrieving the access token")
    os.Exit(1)
  }

  fmt.Println(token, exp)
}
```

## Building Stubbs

*Note: Building a real Stubbs will probably piss off Delos.*

To build the CLI make sure you have [Go](https://golang.org/doc/install) installed and then run the following:

```bash
go build -o stubbs ./cmd/stubbs
```
