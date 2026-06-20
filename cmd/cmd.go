package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/urfave/cli/v3"
)

const BaseURL = "https://inventory.dearsystems.com/ExternalApi/v2"

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:     "account",
		Aliases:  []string{"a"},
		Usage:    "Cin7 account ID",
		Required: true,
		Sources:  cli.EnvVars("CIN7_ACCOUNT_ID"),
	},
	&cli.StringFlag{
		Name:     "key",
		Aliases:  []string{"k"},
		Usage:    "Cin7 application key",
		Required: true,
		Sources:  cli.EnvVars("CIN7_APPLICATION_KEY"),
	},
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "Print request and response details",
	},
}

type Client struct {
	AccountID string
	AppKey    string
	Debug     bool
	http      *http.Client
}

func NewClient(cmd *cli.Command) *Client {
	return &Client{
		AccountID: cmd.String("account"),
		AppKey:    cmd.String("key"),
		Debug:     cmd.Bool("debug"),
		http:      &http.Client{},
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-auth-accountid", c.AccountID)
	req.Header.Set("api-auth-applicationkey", c.AppKey)

	if c.Debug {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Printf("%s\n", dump)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		dump, _ := httputil.DumpResponse(resp, true)
		fmt.Printf("%s\n", dump)
	}

	return resp, nil
}

func PrintSeparator() {
	fmt.Printf("%s\n", strings.Repeat("-", 20))
}
