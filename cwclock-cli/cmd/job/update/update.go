package update

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID            string
	id               string
	name             string
	cron             string
	reportTypes      string
	timePeriod       string
	clientIDs        string
	projectIDs       string
	includeFinancial bool
	enabled          bool
	format           string
)

var fieldFlagNames = []string{
	"name", "cron", "report-types", "time-period", "client-ids", "project-ids",
	"include-financial", "enabled",
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an export job",
	Long:  `This command lets you update an existing export job. Only the flags you pass are changed; everything else (including its targets) keeps its current value. Manage targets with "job target".`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := make(map[string]bool, len(fieldFlagNames))
		for _, flagName := range fieldFlagNames {
			changed[flagName] = cmd.Flags().Changed(flagName)
		}

		fields := handlers.ExportJobFields{
			Name:             name,
			Cron:             cron,
			ReportTypes:      reportTypes,
			TimePeriod:       timePeriod,
			ClientIDs:        clientIDs,
			ProjectIDs:       projectIDs,
			IncludeFinancial: includeFinancial,
			Enabled:          enabled,
		}
		err := handlers.HandleJobUpdate(orgID, id, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	UpdateCmd.DisableFlagsInUseLine = true
	UpdateCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	UpdateCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Job ID to update (required)")
	UpdateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Job name")
	UpdateCmd.Flags().StringVar(&cron, "cron", utils.EMPTY, "Cron expression")
	UpdateCmd.Flags().StringVar(&reportTypes, "report-types", utils.EMPTY, "Comma-separated report types: summary-pdf, summary-csv, detailed-pdf, detailed-csv, unpaid-invoices, all-invoices")
	UpdateCmd.Flags().StringVar(&timePeriod, "time-period", utils.EMPTY, "Time period covered by each run, e.g. now(), now()-1d, now()-1h")
	UpdateCmd.Flags().StringVar(&clientIDs, "client-ids", utils.EMPTY, "Comma-separated client IDs to include (empty = all)")
	UpdateCmd.Flags().StringVar(&projectIDs, "project-ids", utils.EMPTY, "Comma-separated project IDs to include (empty = all)")
	UpdateCmd.Flags().BoolVar(&includeFinancial, "include-financial", false, "Include financial figures in the exported reports")
	UpdateCmd.Flags().BoolVar(&enabled, "enabled", false, "Whether the job is enabled")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
