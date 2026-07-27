package update

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	id       string
	email    string
	name     string
	surname  string
	role     string
	password string
	format   string
)

var fieldFlagNames = []string{"email", "name", "surname", "role", "password"}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user",
	Long:  `This command lets you update an existing user account. Only the flags you pass are changed; everything else keeps its current value. Requires superuser.`,
	Run: func(cmd *cobra.Command, args []string) {
		changed := make(map[string]bool, len(fieldFlagNames))
		for _, flagName := range fieldFlagNames {
			changed[flagName] = cmd.Flags().Changed(flagName)
		}

		fields := handlers.AdminUserFields{
			Email:    email,
			Name:     name,
			Surname:  surname,
			Role:     role,
			Password: password,
		}
		err := handlers.HandleAdminUserUpdate(id, fields, changed, format)
		utils.ExitIfError(err)
	},
}

func init() {
	UpdateCmd.DisableFlagsInUseLine = true
	UpdateCmd.Flags().StringVarP(&id, "id", "i", utils.EMPTY, "User ID or email to update (required)")
	UpdateCmd.Flags().StringVar(&email, "email", utils.EMPTY, "Email")
	UpdateCmd.Flags().StringVar(&name, "name", utils.EMPTY, "First name")
	UpdateCmd.Flags().StringVar(&surname, "surname", utils.EMPTY, "Surname")
	UpdateCmd.Flags().StringVar(&role, "role", utils.EMPTY, "Global role: superuser, confirmed, disabled or ban")
	UpdateCmd.Flags().StringVar(&password, "password", utils.EMPTY, "New password")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", utils.EMPTY, "Output format override: pretty|json")
}
