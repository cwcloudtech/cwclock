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

var validAdminRoles = []string{"superuser", "confirmed", "disabled", "ban"}

type AdminUserFields struct {
	Email    string
	Name     string
	Surname  string
	Role     string
	Password string
}

func adminUserToPayload(user client.AdminUser) client.AdminUserPayload {
	return client.AdminUserPayload{
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
		Role:    user.Role,
	}
}

func mergeAdminUserFields(current client.AdminUser, fields AdminUserFields, changed map[string]bool) client.AdminUserPayload {
	payload := adminUserToPayload(current)

	if changed["email"] {
		payload.Email = fields.Email
	}
	if changed["name"] {
		payload.Name = fields.Name
	}
	if changed["surname"] {
		payload.Surname = fields.Surname
	}
	if changed["role"] {
		payload.Role = fields.Role
	}
	if changed["password"] && utils.IsNotBlank(fields.Password) {
		payload.Password = &fields.Password
	}

	return payload
}

func validateAdminRole(role string) error {
	if !utils.ContainsValue(role, validAdminRoles) {
		return fmt.Errorf("invalid role %q: expected one of %s", role, strings.Join(validAdminRoles, ", "))
	}
	return nil
}

func HandleAdminOrganizationList(formatOverride string) error {
	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	orgs, err := cli.AdminListOrganizations()
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(orgs)
		return nil
	}

	displayAdminOrganizationsAsTable(orgs)
	return nil
}

func HandleAdminOrganizationTransfer(orgOverride string, newOwnerID string, formatOverride string) error {
	trimmedOrgID := strings.TrimSpace(orgOverride)
	if utils.IsBlank(trimmedOrgID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}
	trimmedOwnerID := strings.TrimSpace(newOwnerID)
	if utils.IsBlank(trimmedOwnerID) {
		return fmt.Errorf("new owner id is required: use --owner")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	users, err := cli.AdminListUsers()
	if err != nil {
		return err
	}

	newOwner, found := findAdminUser(users, trimmedOwnerID)
	if !found {
		return fmt.Errorf("user %q not found", trimmedOwnerID)
	}

	org, err := cli.TransferOrganizationOwnership(trimmedOrgID, newOwner.Email)
	if err != nil {
		return err
	}

	renderOrganization(org, formatOverride)
	return nil
}

func HandleAdminUserList(formatOverride string) error {
	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	users, err := cli.AdminListUsers()
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(users)
		return nil
	}

	displayAdminUsersAsTable(users)
	return nil
}

func HandleAdminUserUpdate(id string, fields AdminUserFields, changed map[string]bool, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("user id is required: use -i or --id")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	users, err := cli.AdminListUsers()
	if err != nil {
		return err
	}

	current, found := findAdminUser(users, trimmedID)
	if !found {
		return fmt.Errorf("user %q not found", trimmedID)
	}

	payload := mergeAdminUserFields(current, fields, changed)
	if utils.IsBlank(payload.Email) {
		return fmt.Errorf("email is required: use --email")
	}
	if utils.IsBlank(payload.Name) {
		return fmt.Errorf("name is required: use --name")
	}
	if utils.IsBlank(payload.Surname) {
		return fmt.Errorf("surname is required: use --surname")
	}
	if err := validateAdminRole(payload.Role); err != nil {
		return err
	}

	updated, err := cli.AdminUpdateUser(trimmedID, payload)
	if err != nil {
		return err
	}

	renderAdminUser(updated, formatOverride)
	return nil
}

func HandleAdminUserSetRole(id string, role string, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("user id is required: use -i or --id")
	}
	trimmedRole := strings.TrimSpace(role)
	if err := validateAdminRole(trimmedRole); err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	users, err := cli.AdminListUsers()
	if err != nil {
		return err
	}

	current, found := findAdminUser(users, trimmedID)
	if !found {
		return fmt.Errorf("user %q not found", trimmedID)
	}

	payload := adminUserToPayload(current)
	payload.Role = trimmedRole

	updated, err := cli.AdminUpdateUser(trimmedID, payload)
	if err != nil {
		return err
	}

	renderAdminUser(updated, formatOverride)
	return nil
}

func HandleAdminUserDelete(id string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("user id is required: use -i or --id")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	if err := cli.AdminDeleteUser(trimmedID); err != nil {
		return err
	}

	fmt.Printf("id = %v\n", trimmedID)
	return nil
}

func findAdminUser(users []client.AdminUser, id string) (client.AdminUser, bool) {
	for _, user := range users {
		if user.ID == id {
			return user, true
		}
	}
	return client.AdminUser{}, false
}

func renderAdminUser(user client.AdminUser, formatOverride string) {
	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(user)
		return
	}
	displayAdminUsersAsTable([]client.AdminUser{user})
}

func displayAdminOrganizationsAsTable(orgs []client.AdminOrganization) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Owner Email", "Country", "Currency"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(orgs) {
		table.Append([]string{"No organizations available", "", "", "", ""})
		table.Render()
		return
	}

	for _, org := range orgs {
		table.Append([]string{org.ID, org.Name, org.OwnerEmail, org.Country, org.Currency})
	}
	table.Render()
}

func displayAdminUsersAsTable(users []client.AdminUser) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Email", "Name", "Surname", "Role", "MFA"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(users) {
		table.Append([]string{"No users available", "", "", "", "", ""})
		table.Render()
		return
	}

	for _, user := range users {
		table.Append([]string{
			user.ID, user.Email, user.Name, user.Surname, user.Role,
			strconv.FormatBool(user.MFAEnabled),
		})
	}
	table.Render()
}
