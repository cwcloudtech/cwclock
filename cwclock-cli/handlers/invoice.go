package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func HandleInvoicePreview(orgOverride string, clientID string, projectIDs []string, beginExpr string, endExpr string, toExpr string, outputOverride string) error {
	if utils.IsBlank(strings.TrimSpace(clientID)) {
		return fmt.Errorf("client id is required: use --client")
	}

	start, end, err := resolveDateRangeParams(beginExpr, endExpr, toExpr)
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	resolvedClientID, err := resolveClientID(orgID, clientID)
	if err != nil {
		return err
	}
	resolvedProjects, err := resolveProjects(orgID, projectIDs)
	if err != nil {
		return err
	}
	if err := requireProjectsMatchClients([]string{resolvedClientID}, resolvedProjects); err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	data, filename, err := cli.PreviewInvoice(orgID, client.InvoiceRequest{
		ClientID:       resolvedClientID,
		DateRangeStart: start,
		DateRangeEnd:   end,
		ProjectIDs:     projectIDsOf(resolvedProjects),
	})
	if err != nil {
		return err
	}

	savedPath, err := saveDownloadedFile(data, filename, outputOverride, "invoice.pdf")
	if err != nil {
		return err
	}

	fmt.Printf("Invoice preview saved to %s\n", savedPath)
	return nil
}

func HandleInvoiceGenerate(orgOverride string, clientID string, projectIDs []string, beginExpr string, endExpr string, toExpr string, outputOverride string, formatOverride string) error {
	if utils.IsBlank(strings.TrimSpace(clientID)) {
		return fmt.Errorf("client id is required: use --client")
	}

	start, end, err := resolveDateRangeParams(beginExpr, endExpr, toExpr)
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	resolvedClientID, err := resolveClientID(orgID, clientID)
	if err != nil {
		return err
	}
	resolvedProjects, err := resolveProjects(orgID, projectIDs)
	if err != nil {
		return err
	}
	if err := requireProjectsMatchClients([]string{resolvedClientID}, resolvedProjects); err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	data, filename, err := cli.GenerateInvoice(orgID, client.InvoiceRequest{
		ClientID:       resolvedClientID,
		DateRangeStart: start,
		DateRangeEnd:   end,
		ProjectIDs:     projectIDsOf(resolvedProjects),
	})
	if err != nil {
		return err
	}

	savedPath, err := saveDownloadedFile(data, filename, outputOverride, "invoice.pdf")
	if err != nil {
		return err
	}

	fmt.Printf("Invoice generated and saved to %s\n", savedPath)

	// The generate endpoint streams back the raw PDF, not JSON, so it
	// carries no invoice id. Look the just-created invoice back up (most
	// recent first, see cwclock-api's InvoiceStore.List) to surface it.
	invoices, err := cli.ListInvoices(orgID, resolvedClientID, dayPart(start), dayPart(end))
	if err != nil {
		fmt.Printf("Warning: failed to look up the generated invoice's id: %s\n", err)
		return nil
	}
	if len(invoices) == 0 {
		return nil
	}

	renderInvoice(invoices[0], formatOverride)
	return nil
}

func HandleInvoiceSend(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("invoice id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.SendInvoice(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func HandleInvoiceUpload(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("invoice id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.UploadInvoice(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func HandleInvoiceDelete(orgOverride string, id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("invoice id is required: use -i or --id")
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.DeleteInvoice(orgID, trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func renderInvoice(invoice client.Invoice, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(invoice)
		return
	}
	displayInvoicesAsTable([]client.Invoice{invoice})
}

func displayInvoicesAsTable(invoices []client.Invoice) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Number", "Status", "Total HT", "Total VAT", "Total TTC", "Begin", "End"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(invoices) {
		table.Append([]string{"No invoices available", utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, invoice := range invoices {
		table.Append([]string{
			invoice.ID, invoice.Number, invoice.Status,
			strconv.FormatFloat(invoice.TotalHT, 'f', 2, 64),
			strconv.FormatFloat(invoice.TotalVAT, 'f', 2, 64),
			strconv.FormatFloat(invoice.TotalTTC, 'f', 2, 64),
			invoice.SelectedBeginDate, invoice.SelectedEndDate,
		})
	}
	table.Render()
}
