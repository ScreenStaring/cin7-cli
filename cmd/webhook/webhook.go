package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cheynewallace/tabby"
	"github.com/sshaw/cin7/cmd"
	"github.com/urfave/cli/v3"
)

var Cmd cli.Command

type ExternalHeader struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type Webhook struct {
	ID                        string           `json:"ID"`
	Type                      string           `json:"Type"`
	Name                      string           `json:"Name"`
	IsActive                  bool             `json:"IsActive"`
	ExternalURL               string           `json:"ExternalURL"`
	ExternalAuthorizationType string           `json:"ExternalAuthorizationType"`
	ExternalUserName          *string          `json:"ExternalUserName"`
	ExternalPassword          *string          `json:"ExternalPassword"`
	ExternalBearerToken       *string          `json:"ExternalBearerToken"`
	ExternalHeaders           []ExternalHeader `json:"ExternalHeaders"`
}

type listResponse struct {
	Webhooks []Webhook `json:"Webhooks"`
	Total    int       `json:"Total"`
}

func strPtr(s string) *string { return &s }

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fetchWebhooks(ctx context.Context, client *cmd.Client) ([]Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/webhooks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed (%s): %s", resp.Status, body)
	}

	var result listResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Webhooks, nil
}

func listAction(ctx context.Context, c *cli.Command) error {
	client := cmd.NewClient(c)

	webhooks, err := fetchWebhooks(ctx, client)
	if err != nil {
		return err
	}

	t := tabby.New()
	for i, w := range webhooks {
		t.AddLine("ID", w.ID)
		t.AddLine("Name", w.Name)
		t.AddLine("Type", w.Type)
		t.AddLine("URL", w.ExternalURL)
		t.AddLine("Auth Type", w.ExternalAuthorizationType)
		t.AddLine("External User", derefStr(w.ExternalUserName))
		t.AddLine("External Password", derefStr(w.ExternalPassword))
		t.AddLine("External Token", derefStr(w.ExternalBearerToken))
		t.AddLine("Headers", w.ExternalHeaders)
		t.AddLine("Active", w.IsActive)
		t.Print()
		if i < len(webhooks)-1 {
			cmd.PrintSeparator()
		}
	}

	return nil
}

func deleteAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("you must supply a webhook ID")
	}

	id := c.Args().Get(0)
	client := cmd.NewClient(c)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, cmd.BaseURL+"/webhooks?ID="+id, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed (%s): %s", resp.Status, body)
	}

	fmt.Printf("Webhook %s deleted\n", id)
	return nil
}

func createAction(ctx context.Context, c *cli.Command) error {
	payload := Webhook{
		Type:                      c.String("type"),
		Name:                      c.String("name"),
		IsActive:                  true,
		ExternalURL:               c.String("address"),
		ExternalAuthorizationType: c.String("auth"),
		ExternalBearerToken:       strPtr(c.String("token")),
		ExternalUserName:          strPtr(c.String("user")),
		ExternalPassword:          strPtr(c.String("password")),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := cmd.NewClient(c)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cmd.BaseURL+"/webhooks", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed (%s): %s", resp.Status, body)
	}

	var created Webhook
	if err := json.Unmarshal(body, &created); err != nil {
		return err
	}

	fmt.Printf("Webhook created: %s\n", created.ID)
	return nil
}

func updateAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("you must supply a webhook ID")
	}

	id := c.Args().Get(0)
	client := cmd.NewClient(c)

	webhooks, err := fetchWebhooks(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to fetch webhook: %w", err)
	}

	var existing *Webhook
	for i := range webhooks {
		if webhooks[i].ID == id {
			existing = &webhooks[i]
			break
		}
	}
	if existing == nil {
		return fmt.Errorf("webhook %s not found", id)
	}

	if c.IsSet("type") {
		existing.Type = c.String("type")
	}
	if c.IsSet("address") {
		existing.ExternalURL = c.String("address")
	}
	if c.IsSet("auth") {
		existing.ExternalAuthorizationType = c.String("auth")
	}
	if c.IsSet("token") {
		if s := c.String("token"); s != "" {
			existing.ExternalBearerToken = strPtr(s)
		} else {
			existing.ExternalBearerToken = nil
		}
	}
	if c.IsSet("user") {
		if s := c.String("user"); s != "" {
			existing.ExternalUserName = strPtr(s)
		} else {
			existing.ExternalUserName = nil
		}
	}
	if c.IsSet("password") {
		if s := c.String("password"); s != "" {
			existing.ExternalPassword = strPtr(s)
		} else {
			existing.ExternalPassword = nil
		}
	}

	data, err := json.Marshal(existing)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, cmd.BaseURL+"/webhooks", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed (%s): %s", resp.Status, body)
	}

	fmt.Printf("Webhook %s updated\n", id)
	return nil
}

func init() {
	sharedFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t"},
			Usage:   "Webhook type (e.g. Sale/OrderAuthorised)",
		},
		&cli.StringFlag{
			Name:  "address",
			Usage: "External URL",
		},
		&cli.StringFlag{
			Name:  "auth",
			Usage: "External authorization type",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "External bearer token",
		},
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "External username",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"p"},
			Usage:   "External password",
		},
	}

	Cmd = cli.Command{
		Name:    "webhook",
		Aliases: []string{"webhooks", "hooks", "w"},
		Usage:   "Webhook utilities",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List webhooks for the account",
				Flags:   cmd.Flags,
				Action:  listAction,
			},
			{
				Name:      "delete",
				Aliases:   []string{"del", "rm", "d"},
				Usage:     "Delete the webhook with the given ID",
				ArgsUsage: "<ID>",
				Flags:     cmd.Flags,
				Action:    deleteAction,
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a webhook",
				Flags: append(
					cmd.Flags,
					append(sharedFlags, &cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Webhook name",
					})...),
				Action: createAction,
			},
			{
				Name:      "update",
				Aliases:   []string{"u"},
				Usage:     "Update a webhook",
				ArgsUsage: "<ID>",
				Flags:     append(cmd.Flags, sharedFlags...),
				Action:    updateAction,
			},
		},
	}
}
