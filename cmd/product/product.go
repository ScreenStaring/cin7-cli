package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/sshaw/cin7/cmd"
	"github.com/urfave/cli/v3"
)

var Cmd cli.Command

type Movement struct {
	TaskID     string  `json:"TaskID"`
	Type       string  `json:"Type"`
	Date       string  `json:"Date"`
	Number     string  `json:"Number"`
	Status     string  `json:"Status"`
	Quantity   float64 `json:"Quantity"`
	Amount     float64 `json:"Amount"`
	Location   string  `json:"Location"`
	BatchSN    string  `json:"BatchSN"`
	ExpiryDate string  `json:"ExpiryDate"`
	FromTo     string  `json:"FromTo"`
}

type Product struct {
	ID                  string     `json:"ID"`
	SKU                 string     `json:"SKU"`
	Barcode             string     `json:"Barcode"`
	Name                string     `json:"Name"`
	Category            string     `json:"Category"`
	Type                string     `json:"Type"`
	Status              string     `json:"Status"`
	Tags                string     `json:"Tags"`
	Sellable            bool       `json:"Sellable"`
	DefaultLocation     string     `json:"DefaultLocation"`
	CostingMethod       string     `json:"CostingMethod"`
	Description         string     `json:"Description"`
	AverageCost         float64    `json:"AverageCost"`
	UOM                 string     `json:"UOM"`
	HSCode              string     `json:"HSCode"`
	CountryOfOriginCode string     `json:"CountryOfOriginCode"`
	LastModifiedOn      string     `json:"LastModifiedOn"`
	Movements           []Movement `json:"Movements"`
}

type listResponse struct {
	Products []Product `json:"Products"`
	Total    int       `json:"Total"`
}

func trimDate(s string) string {
	if i := strings.IndexByte(s, 'T'); i != -1 {
		return s[:i]
	}
	return s
}

func listAction(ctx context.Context, c *cli.Command) error {
	params := url.Values{
		"Page":  {strconv.Itoa(c.Int("page"))},
		"Limit": {strconv.Itoa(c.Int("limit"))},
	}
	includes := c.StringSlice("include")
	if slices.Contains(includes, "movements") {
		params.Set("IncludeMovements", "true")
	}

	if arg := c.Args().Get(0); arg != "" {
		if v, ok := strings.CutPrefix(arg, "sku="); ok {
			params.Set("Sku", v)
		} else if v, ok := strings.CutPrefix(arg, "name="); ok {
			params.Set("Name", v)
		} else {
			params.Set("ID", arg)
		}
	}

	client := cmd.NewClient(c)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/product?"+params.Encode(), nil)
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

	var result listResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	t := tabby.New()
	for i, p := range result.Products {
		t.AddLine("ID", p.ID)
		t.AddLine("SKU", p.SKU)
		t.AddLine("Barcode", p.Barcode)
		t.AddLine("Name", p.Name)
		t.AddLine("Category", p.Category)
		t.AddLine("Type", p.Type)
		t.AddLine("Status", p.Status)
		t.AddLine("Tags", p.Tags)
		t.AddLine("Sellable", p.Sellable)
		t.AddLine("Default Location", p.DefaultLocation)
		t.AddLine("Costing Method", p.CostingMethod)
		t.AddLine("Description", p.Description)
		t.AddLine("Average Cost", p.AverageCost)
		t.AddLine("UOM", p.UOM)
		t.AddLine("HS Code", p.HSCode)
		t.AddLine("Country of Origin", p.CountryOfOriginCode)
		t.AddLine("Last Modified", p.LastModifiedOn)
		t.Print()

		if len(p.Movements) > 0 {
			slices.SortFunc(p.Movements, func(a, b Movement) int {
				return strings.Compare(b.Date, a.Date)
			})
			fmt.Println("Movements")
			mt := tabby.New()
			mt.AddHeader("Type", "Date", "Number", "Status", "Quantity", "Amount", "Location", "Batch/SN", "Expiry Date", "From/To")
			for _, m := range p.Movements {
				mt.AddLine(m.Type, trimDate(m.Date), m.Number, m.Status, m.Quantity, m.Amount, m.Location, m.BatchSN, trimDate(m.ExpiryDate), m.FromTo)
			}
			mt.Print()
			fmt.Println()
		}

		if i < len(result.Products)-1 {
			cmd.PrintSeparator()
		}
	}

	return nil
}

func init() {
	Cmd = cli.Command{
		Name:    "product",
		Aliases: []string{"products"},
		Usage:   "Product utilities",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Aliases:   []string{"ls"},
				Usage:     "List products",
				ArgsUsage: "[ID | sku=<value> | name=<value>]",
				Flags: append(cmd.Flags,
					&cli.IntFlag{
						Name:    "page",
						Aliases: []string{"p"},
						Usage:   "Page number",
						Value:   1,
					},
					&cli.StringSliceFlag{
						Name:    "include",
						Aliases: []string{"i"},
						Usage:   "Include additional data: movements",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Limit number of results",
					},
				),
				Action: listAction,
			},
		},
	}
}
