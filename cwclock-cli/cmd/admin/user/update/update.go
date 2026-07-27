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
	UpdateCmd.Flags().StringVarP(&id, "id", "i", "", "User ID to update (required)")
	UpdateCmd.Flags().StringVar(&email, "email", "", "Email")
	UpdateCmd.Flags().StringVar(&name, "name", "", "First name")
	UpdateCmd.Flags().StringVar(&surname, "surname", "", "Surname")
	UpdateCmd.Flags().StringVar(&role, "role", "", "Global role: superuser, confirmed, disabled or ban")
	UpdateCmd.Flags().StringVar(&password, "password", "", "New password")
	UpdateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
