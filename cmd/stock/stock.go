package stock

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

// Stock adjustments

type StockLine struct {
	ProductID    string  `json:"ProductID"`
	SKU          string  `json:"SKU"`
	ProductName  string  `json:"ProductName"`
	Quantity     float64 `json:"Quantity"`
	UnitCost     float64 `json:"UnitCost"`
	LocationID   string  `json:"LocationID"`
	Location     string  `json:"Location"`
	BatchSN      string  `json:"BatchSN"`
	ExpiryDate   string  `json:"ExpiryDate"`
	ReceivedDate string  `json:"ReceivedDate"`
	Comments     string  `json:"Comments"`
}

type Transaction struct {
	ID            string  `json:"ID"`
	Debit         string  `json:"Debit"`
	Credit        string  `json:"Credit"`
	Amount        float64 `json:"Amount"`
	EffectiveDate string  `json:"EffectiveDate"`
}

type StockAdjustment struct {
	TaskID             string        `json:"TaskID"`
	EffectiveDate      string        `json:"EffectiveDate"`
	StocktakeNumber    string        `json:"StocktakeNumber"`
	Status             string        `json:"Status"`
	Account            string        `json:"Account"`
	Reference          string        `json:"Reference"`
	Comment            string        `json:"Comment"`
	ExistingStockLines []StockLine   `json:"ExistingStockLines"`
	NewStockLines      []StockLine   `json:"NewStockLines"`
	Transactions       []Transaction `json:"Transactions"`
}

type adjListResponse struct {
	Total               int               `json:"Total"`
	Page                int               `json:"Page"`
	StockAdjustmentList []StockAdjustment `json:"StockAdjustmentList"`
}

// Stock takes

type StocktakeCategory struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

type StocktakeListItem struct {
	TaskID          string `json:"TaskID"`
	EffectiveDate   string `json:"EffectiveDate"`
	StocktakeNumber string `json:"StocktakeNumber"`
	Status          string `json:"Status"`
	Account         string `json:"Account"`
	LocationID      string `json:"LocationID"`
	Location        string `json:"Location"`
	Tags            string `json:"Tags"`
	PickZones       string `json:"PickZones"`
	StockLocators   string `json:"StockLocators"`
	Categories      string `json:"Categories"`
	Brands          string `json:"Brands"`
	Bins            string `json:"Bins"`
	Reference       string `json:"Reference"`
	CreatedDate     string `json:"CreatedDate"`
	LastUpdatedDate string `json:"LastUpdatedDate"`
	LastUpdatedBy   string `json:"LastUpdatedBy"`
}

type Stocktake struct {
	TaskID          string              `json:"TaskID"`
	EffectiveDate   string              `json:"EffectiveDate"`
	StocktakeNumber string              `json:"StocktakeNumber"`
	Status          string              `json:"Status"`
	Account         string              `json:"Account"`
	LocationID      string              `json:"LocationID"`
	Location        string              `json:"Location"`
	Tags            []string            `json:"Tags"`
	Categories      []StocktakeCategory `json:"Categories"`
	Bins            json.RawMessage     `json:"Bins"`
	Reference       string              `json:"Reference"`
}

type takeListResponse struct {
	Total               int                 `json:"Total"`
	Page                int                 `json:"Page"`
	StockAdjustmentList []StocktakeListItem `json:"StockAdjustmentList"`
}

func trimDate(s string) string {
	if i := strings.IndexByte(s, 'T'); i != -1 {
		return s[:i]
	}
	return s
}

func printStockLines(label string, lines []StockLine) {
	if len(lines) == 0 {
		return
	}
	fmt.Println(label)
	t := tabby.New()
	t.AddHeader("SKU", "Product Name", "Qty", "Unit Cost", "Location", "Batch/SN", "Expiry Date", "Received Date", "Comments")
	for _, sl := range lines {
		t.AddLine(sl.SKU, sl.ProductName, sl.Quantity, sl.UnitCost, sl.Location, sl.BatchSN, trimDate(sl.ExpiryDate), trimDate(sl.ReceivedDate), sl.Comments)
	}
	t.Print()
	fmt.Println()
}

func printTransactions(txns []Transaction) {
	if len(txns) == 0 {
		return
	}
	fmt.Println("Transactions")
	t := tabby.New()
	t.AddHeader("ID", "Debit", "Credit", "Amount", "Effective Date")
	for _, tx := range txns {
		t.AddLine(tx.ID, tx.Debit, tx.Credit, tx.Amount, trimDate(tx.EffectiveDate))
	}
	t.Print()
	fmt.Println()
}

func printAdjustmentProps(sa StockAdjustment) {
	t := tabby.New()
	t.AddLine("Task ID", sa.TaskID)
	t.AddLine("Effective Date", trimDate(sa.EffectiveDate))
	t.AddLine("Stocktake Number", sa.StocktakeNumber)
	t.AddLine("Status", sa.Status)
	t.AddLine("Account", sa.Account)
	t.AddLine("Reference", sa.Reference)
	t.AddLine("Comment", sa.Comment)
	t.Print()
	fmt.Println()
}

func adjustmentsAction(ctx context.Context, c *cli.Command) error {
	client := cmd.NewClient(c)

	if taskID := c.Args().Get(0); taskID != "" {
		params := url.Values{}
		params.Set("TaskID", taskID)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/stockadjustment?"+params.Encode(), nil)
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

		var sa StockAdjustment
		if err := json.Unmarshal(body, &sa); err != nil {
			return err
		}

		printAdjustmentProps(sa)
		printStockLines("Existing Stock Lines", sa.ExistingStockLines)
		printStockLines("New Stock Lines", sa.NewStockLines)
		printTransactions(sa.Transactions)
		return nil
	}

	params := url.Values{
		"Page":  {strconv.Itoa(c.Int("page"))},
		"Limit": {strconv.Itoa(c.Int("limit"))},
	}
	if status := c.String("status"); status != "" {
		params.Set("Status", status)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/stockadjustmentList?"+params.Encode(), nil)
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

	var result adjListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	slices.SortFunc(result.StockAdjustmentList, func(a, b StockAdjustment) int {
		return strings.Compare(b.EffectiveDate, a.EffectiveDate)
	})

	t := tabby.New()
	for i, sa := range result.StockAdjustmentList {
		t.AddLine("Task ID", sa.TaskID)
		t.AddLine("Effective Date", trimDate(sa.EffectiveDate))
		t.AddLine("Stocktake Number", sa.StocktakeNumber)
		t.AddLine("Status", sa.Status)
		t.AddLine("Account", sa.Account)
		t.AddLine("Reference", sa.Reference)
		t.AddLine("Comment", sa.Comment)
		t.Print()

		if i < len(result.StockAdjustmentList)-1 {
			cmd.PrintSeparator()
		}
	}

	return nil
}

func takeAction(ctx context.Context, c *cli.Command) error {
	client := cmd.NewClient(c)

	if taskID := c.Args().Get(0); taskID != "" {
		params := url.Values{}
		params.Set("TaskID", taskID)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/stocktake?"+params.Encode(), nil)
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

		var st Stocktake
		if err := json.Unmarshal(body, &st); err != nil {
			return err
		}

		t := tabby.New()
		t.AddLine("Task ID", st.TaskID)
		t.AddLine("Effective Date", trimDate(st.EffectiveDate))
		t.AddLine("Stocktake Number", st.StocktakeNumber)
		t.AddLine("Status", st.Status)
		t.AddLine("Account", st.Account)
		t.AddLine("Location ID", st.LocationID)
		t.AddLine("Location", st.Location)
		t.AddLine("Reference", st.Reference)
		t.Print()
		fmt.Println()

		if len(st.Tags) > 0 {
			fmt.Println("Tags")
			tt := tabby.New()
			tt.AddHeader("Tag")
			for _, tag := range st.Tags {
				tt.AddLine(tag)
			}
			tt.Print()
			fmt.Println()
		}

		if len(st.Categories) > 0 {
			fmt.Println("Categories")
			ct := tabby.New()
			ct.AddHeader("ID", "Name")
			for _, cat := range st.Categories {
				ct.AddLine(cat.ID, cat.Name)
			}
			ct.Print()
			fmt.Println()
		}

		return nil
	}

	params := url.Values{
		"Page":  {strconv.Itoa(c.Int("page"))},
		"Limit": {strconv.Itoa(c.Int("limit"))},
	}
	if status := c.String("status"); status != "" {
		params.Set("Status", status)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cmd.BaseURL+"/stockTakeList?"+params.Encode(), nil)
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

	var result takeListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	slices.SortFunc(result.StockAdjustmentList, func(a, b StocktakeListItem) int {
		return strings.Compare(b.EffectiveDate, a.EffectiveDate)
	})

	t := tabby.New()
	for i, st := range result.StockAdjustmentList {
		t.AddLine("Task ID", st.TaskID)
		t.AddLine("Effective Date", trimDate(st.EffectiveDate))
		t.AddLine("Stocktake Number", st.StocktakeNumber)
		t.AddLine("Status", st.Status)
		t.AddLine("Account", st.Account)
		t.AddLine("Location ID", st.LocationID)
		t.AddLine("Location", st.Location)
		t.AddLine("Tags", st.Tags)
		t.AddLine("Pick Zones", st.PickZones)
		t.AddLine("Stock Locators", st.StockLocators)
		t.AddLine("Categories", st.Categories)
		t.AddLine("Brands", st.Brands)
		t.AddLine("Bins", st.Bins)
		t.AddLine("Reference", st.Reference)
		t.AddLine("Created Date", trimDate(st.CreatedDate))
		t.AddLine("Last Updated Date", trimDate(st.LastUpdatedDate))
		t.AddLine("Last Updated By", st.LastUpdatedBy)
		t.Print()

		if i < len(result.StockAdjustmentList)-1 {
			cmd.PrintSeparator()
		}
	}

	return nil
}

var stockFlags = append(cmd.Flags,
	&cli.IntFlag{
		Name:    "page",
		Aliases: []string{"p"},
		Usage:   "Page number",
		Value:   1,
	},
	&cli.StringFlag{
		Name:    "status",
		Aliases: []string{"s"},
		Usage:   "Filter by status",
	},
	&cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "Limit number of results",
	},
)

func init() {
	Cmd = cli.Command{
		Name:  "stock",
		Usage: "Stock utilities",
		Commands: []*cli.Command{
			{
				Name:    "adjustments",
				Aliases: []string{"a", "adj"},
				Usage:   "Stock adjustment commands",
				Commands: []*cli.Command{
					{
						Name:      "list",
						Aliases:   []string{"ls"},
						Usage:     "List stock adjustments or show detail for a given TaskID",
						ArgsUsage: "[TaskID]",
						Flags:     stockFlags,
						Action:    adjustmentsAction,
					},
				},
			},
			{
				Name:    "take",
				Aliases: []string{"t"},
				Usage:   "Stock take commands",
				Commands: []*cli.Command{
					{
						Name:      "list",
						Aliases:   []string{"ls"},
						Usage:     "List stock takes or show detail for a given TaskID",
						ArgsUsage: "[TaskID]",
						Flags:     stockFlags,
						Action:    takeAction,
					},
				},
			},
		},
	}
}
