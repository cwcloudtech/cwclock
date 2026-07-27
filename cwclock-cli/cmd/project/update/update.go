package update

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID        string
	id           string
	clientID     string
	name         string
	color        string
	dailyRate    float64
	subdivisions string
	format       string
)

var fieldFlagNames = []string{"name", "color", "daily-rate", "subdivisions"}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a project",
	Long:  `This command lets you update an existing project. Only the flags you pass are changed; everything else keeps its current value.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := make(map[string]bool, len(fieldFlagNames))
		for _, flagName := range fieldFlagNames {
			changed[flagName] = cmd.Flags().Changed(flagName)
		}

		fields := handlers.ProjectFields{
			Name:         name,
			Color:        color,
			DailyRate:    dailyRate,
			Subdivisions: subdivisions,
		}
		err := handlers.HandleProjectUpdate(orgID, id, clientID, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	UpdateCmd.DisableFlagsInUseLine = true
	UpdateCmd.Flags().StringVarP(&orgID, "org", "o", utils.EMPTY, "Organization ID (overrides configured org_id)")
	UpdateCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "Project ID to update (required)")
	UpdateCmd.Flags().StringVarP(&clientID, "client", "c", utils.EMPTY, "Reassign to this client ID")
	UpdateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "Project name")
	UpdateCmd.Flags().StringVar(&color, "color", utils.EMPTY, "Project color")
	UpdateCmd.Flags().Float64Var(&dailyRate, "daily-rate", 0, "Daily rate")
	UpdateCmd.Flags().StringVar(&subdivisions, "subdivisions", utils.EMPTY, "Comma-separated list of subdivisions")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
