package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/m-rots/stubbs"
)

var cli struct {
	Path     string   `arg:"" name:"path" help:"Service Account file path" type:"path"`
	Scopes   []string `required:"" name:"scope" short:"s" help:"Authentication scopes"`
	Terse    bool     `optional:"" name:"terse" short:"t" help:"Only show the access token"`
	Lifetime int      `optional:"" default:"3600" name:"lifetime" short:"l" help:"Token lifetime in seconds (max. 3600)"`
}

type googleServiceAccount struct {
	Email      string `json:"client_email"`
	PrivateKey string `json:"private_key"`
}

func main() {
	ctx := kong.Parse(&cli, kong.Name("stubbs"), kong.UsageOnError())

	if cli.Lifetime > 3600 {
		fmt.Println("Maximum lifetime is 3600 seconds")
		os.Exit(1)
	}

	if ctx.Command() != "<path>" {
		fmt.Println("Unknown command")
		os.Exit(1)
	}

	file, err := os.Open(cli.Path)
	if err != nil {
		fmt.Println("Could not open: " + cli.Path)
		os.Exit(1)
	}

	decoder := json.NewDecoder(file)
	sa := new(googleServiceAccount)

	if decoder.Decode(sa) != nil {
		fmt.Println("Error decoding service account")
		os.Exit(1)
	}

	priv, err := stubbs.ParseKey(sa.PrivateKey)
	if err != nil {
		fmt.Println("Invalid private key")
		os.Exit(1)
	}

	scopes := mapScopes(cli.Scopes)
	account := stubbs.New(sa.Email, &priv, scopes, cli.Lifetime)

	token, exp, err := account.AccessToken()
	if err != nil {
		fmt.Println("Error retrieving the access token")
		os.Exit(1)
	}

	if cli.Terse {
		fmt.Println(token)
		os.Exit(0)
	}

	fmt.Printf("Email:\n%v\n\n", sa.Email)
	fmt.Printf("Scopes:\n%v\n\n", strings.Join(scopes, "\n"))

	fmt.Printf("Access Token:\n%v\n\n", token)
	setupCloseHandler()

	fmt.Print("Status:\n")

	for {
		validFor := exp - time.Now().Unix()
		if validFor < 1 {
			break
		}

		fmt.Printf("%vValid for another %v seconds", clear, validFor)
		time.Sleep(1 * time.Second)
	}

	fmt.Print(clear + "Expired\n")
}

func isValidURL(scope string) bool {
	_, err := url.ParseRequestURI(scope)
	if err != nil {
		return false
	}

	u, err := url.Parse(scope)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func mapScopes(scopes []string) []string {
	var fullyScoped []string

	for _, scope := range scopes {
		if isValidURL(scope) {
			fullyScoped = append(fullyScoped, scope)
		} else {
			fullyScoped = append(fullyScoped, "https://www.googleapis.com/auth/"+scope)
		}
	}

	return fullyScoped
}

const clear = "\r\033[K"

func setupCloseHandler() {
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-termChan

		fmt.Print(clear)
		os.Exit(0)
	}()
}
