package main

import (
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
	Lifetime int64    `optional:"" default:"3600" name:"lifetime" short:"l" help:"Token lifetime in seconds (max. 3600)"`
}

func main() {
	ctx := kong.Parse(&cli, kong.Name("stubbs"), kong.UsageOnError())

	if cli.Lifetime < 1 {
		fmt.Println("Minimum lifetime is 1 second")
	}

	if cli.Lifetime > 3600 {
		fmt.Println("Maximum lifetime is 3600 seconds")
		os.Exit(1)
	}

	if ctx.Command() != "<path>" {
		fmt.Println("Unknown command")
		os.Exit(1)
	}

	scopes := mapScopes(cli.Scopes)
	account, err := stubbs.FromFile(cli.Path, scopes, stubbs.WithLifetime(cli.Lifetime, 0))
	if err != nil {
		fmt.Println("Could not open: " + cli.Path)
		os.Exit(1)
	}

	token, exp, err := account.AccessToken()
	if err != nil {
		fmt.Println("Error retrieving the access token")
		os.Exit(1)
	}

	if cli.Terse {
		fmt.Println(token)
		os.Exit(0)
	}

	fmt.Printf("Email:\n%v\n\n", account.Email())
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
