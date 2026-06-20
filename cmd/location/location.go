package location

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/sshaw/cin7/cmd"
	"github.com/urfave/cli/v3"
)

var Cmd cli.Command

type Bin struct {
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	IsDeprecated bool   `json:"IsDeprecated"`
	IsStaging    bool   `json:"IsStaging"`
}

type Location struct {
	ID                   string          `json:"ID"`
	Name                 string          `json:"Name"`
	IsDefault            bool            `json:"IsDefault"`
	AddressLine1         string          `json:"AddressLine1"`
	AddressLine2         string          `json:"AddressLine2"`
	AddressCitySuburb    string          `json:"AddressCitySuburb"`
	AddressStateProvince string          `json:"AddressStateProvince"`
	AddressZipPostCode   string          `json:"AddressZipPostCode"`
	AddressCountry       string          `json:"AddressCountry"`
	Bins                 []Bin           `json:"Bins"`
	PickZones            json.RawMessage `json:"PickZones"`
	IsShopfloor          bool            `json:"IsShopfloor"`
	IsCoMan              bool            `json:"IsCoMan"`
	IsStaging            bool            `json:"IsStaging"`
	FixedAssetsLocation  bool            `json:"FixedAssetsLocation"`
	ReferenceCount       int             `json:"ReferenceCount"`
	ParentID             string          `json:"ParentID"`
	Deprecated           bool            `json:"IsDeprecated"`
}

type listResponse struct {
	LocationList []Location `json:"LocationList"`
	Total        int        `json:"Total"`
}

func rawJSON(r json.RawMessage) string {
	if r == nil || string(r) == "null" {
		return ""
	}
	return string(r)
}

func listAction(ctx context.Context, c *cli.Command) error {
	params := url.Values{
		"Page":  {strconv.Itoa(c.Int("page"))},
		"Limit": {strconv.Itoa(c.Int("limit"))},
	}

	if arg := c.Args().Get(0); arg != "" {
		if v, ok := strings.CutPrefix(arg, "name="); ok {
			params.Set("Name", v)
		} else {
			params.Set("ID", arg)
		}
	}

	client := cmd.NewClient(c)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/ref/location?"+params.Encode(), nil)
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
	for i, l := range result.LocationList {
		t.AddLine("ID", l.ID)
		t.AddLine("Name", l.Name)
		t.AddLine("Default", l.IsDefault)
		t.AddLine("Address Line 1", l.AddressLine1)
		t.AddLine("Address Line 2", l.AddressLine2)
		t.AddLine("City/Suburb", l.AddressCitySuburb)
		t.AddLine("State/Province", l.AddressStateProvince)
		t.AddLine("Zip/Post Code", l.AddressZipPostCode)
		t.AddLine("Country", l.AddressCountry)
		t.AddLine("Pick Zones", rawJSON(l.PickZones))
		t.AddLine("Shopfloor", l.IsShopfloor)
		t.AddLine("CoMan", l.IsCoMan)
		t.AddLine("Staging", l.IsStaging)
		t.AddLine("Fixed Assets Location", l.FixedAssetsLocation)
		t.AddLine("Reference Count", l.ReferenceCount)
		t.AddLine("Parent ID", l.ParentID)
		t.AddLine("Deprecated", l.Deprecated)
		t.Print()

		fmt.Println("Bins")

		if len(l.Bins) > 0 {
			bt := tabby.New()
			bt.AddHeader("ID", "Name", "Deprecated", "Staging")
			for _, b := range l.Bins {
				bt.AddLine(b.ID, b.Name, b.IsDeprecated, b.IsStaging)
			}
			bt.Print()
			fmt.Println()
		}

		if i < len(result.LocationList)-1 {
			cmd.PrintSeparator()
		}
	}

	return nil
}

func init() {
	Cmd = cli.Command{
		Name:    "location",
		Aliases: []string{"locations", "loc"},
		Usage:   "Location utilities",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Aliases:   []string{"ls"},
				Usage:     "List locations",
				ArgsUsage: "[ID | name=<value>]",
				Flags: append(cmd.Flags,
					&cli.IntFlag{
						Name:    "page",
						Aliases: []string{"p"},
						Usage:   "Page number",
						Value:   1,
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
