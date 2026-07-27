package create

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID        string
	clientID     string
	name         string
	color        string
	dailyRate    float64
	subdivisions string
	format       string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project",
	Long:  `This command lets you create a new project for a client.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := map[string]bool{
			"daily-rate": cmd.Flags().Changed("daily-rate"),
		}

		fields := handlers.ProjectFields{
			Name:         name,
			Color:        color,
			DailyRate:    dailyRate,
			Subdivisions: subdivisions,
		}
		err := handlers.HandleProjectCreate(orgID, clientID, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	CreateCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Client ID (required)")
	CreateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Project name (required)")
	CreateCmd.Flags().StringVar(&color, "color", utils.EMPTY, "Project color")
	CreateCmd.Flags().Float64Var(&dailyRate, "daily-rate", 0, "Daily rate")
	CreateCmd.Flags().StringVar(&subdivisions, "subdivisions", utils.EMPTY, "Comma-separated list of subdivisions")
	CreateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
