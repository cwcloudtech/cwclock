package delete

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID  string
	id     string
	offset int
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a target from an export job",
	Long:  `This command lets you remove a target from an export job by its offset in the targets array (see "job target ls").`,
	Run: func(cmd *cobra.Command, args []string) {
		err := handlers.HandleJobTargetDelete(orgID, id, offset)
		utils.ExitIfError(err)
	},
}

func init() {
	DeleteCmd.DisableFlagsInUseLine = true
	DeleteCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	DeleteCmd.Flags().StringVarP(&id, "id", "i", "", "Job ID (required)")
	DeleteCmd.Flags().IntVar(&offset, "offset", -1, "Offset of the target to remove in the targets array (required)")
}
