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

// OrgConnectionFields is the flag-facing shape of an organization's
// external connection. Unlike an export job target, an organization
// connection has no "email" type - only s3/google_drive/git.
type OrgConnectionFields struct {
	Type                    string
	Endpoint                string
	BucketName              string
	Region                  string
	AccessKey               string
	SecretKey               string
	ServiceAccountBase64    string
	FolderID                string
	RepoURL                 string
	Username                string
	Password                string
	SSHPrivateKey           string
	SSHPrivateKeyPassphrase string
	Path                    string
	FlatDirectory           bool
}

func (f OrgConnectionFields) hasAny() bool {
	return utils.IsNotBlank(f.Type) || utils.IsNotBlank(f.Endpoint) || utils.IsNotBlank(f.BucketName) ||
		utils.IsNotBlank(f.Region) || utils.IsNotBlank(f.AccessKey) || utils.IsNotBlank(f.SecretKey) ||
		utils.IsNotBlank(f.ServiceAccountBase64) || utils.IsNotBlank(f.FolderID) ||
		utils.IsNotBlank(f.RepoURL) || utils.IsNotBlank(f.Username) || utils.IsNotBlank(f.Password) ||
		utils.IsNotBlank(f.SSHPrivateKey) || utils.IsNotBlank(f.SSHPrivateKeyPassphrase)
}

func (f OrgConnectionFields) toConnection() (client.ExternalConnection, error) {
	connType := strings.TrimSpace(f.Type)
	switch connType {
	case "s3", "google_drive", "git":
	case "":
		return client.ExternalConnection{}, fmt.Errorf("--type is required: expected s3, google_drive or git")
	default:
		return client.ExternalConnection{}, fmt.Errorf("invalid type %q: expected s3, google_drive or git", connType)
	}

	return client.ExternalConnection{
		Type:                    connType,
		Endpoint:                f.Endpoint,
		BucketName:              f.BucketName,
		Region:                  f.Region,
		AccessKey:               f.AccessKey,
		SecretKey:               f.SecretKey,
		ServiceAccountBase64:    f.ServiceAccountBase64,
		FolderID:                f.FolderID,
		RepoURL:                 f.RepoURL,
		Username:                f.Username,
		Password:                f.Password,
		SSHPrivateKey:           f.SSHPrivateKey,
		SSHPrivateKeyPassphrase: f.SSHPrivateKeyPassphrase,
		Path:                    f.Path,
		FlatDirectory:           f.FlatDirectory,
	}, nil
}

func HandleOrgConnectionList(id string, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	org, err := cli.GetOrganization(trimmedID)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(org.ExternalConnections)
		return nil
	}

	displayExternalConnectionsAsTable(org.ExternalConnections)
	return nil
}

func HandleOrgConnectionCreate(id string, fields OrgConnectionFields, formatOverride string) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}
	if !fields.hasAny() {
		return fmt.Errorf("at least one connection field is required: use --type s3/google_drive/git with its connection fields")
	}

	conn, err := fields.toConnection()
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	org, err := cli.AddOrganizationExternalConnection(trimmedID, conn)
	if err != nil {
		return err
	}

	if config.GetDefaultFormat(formatOverride) == "json" {
		utils.PrintJson(org.ExternalConnections)
		return nil
	}

	displayExternalConnectionsAsTable(org.ExternalConnections)
	return nil
}

func HandleOrgConnectionDelete(id string, offset int) error {
	trimmedID := strings.TrimSpace(id)
	if utils.IsBlank(trimmedID) {
		return fmt.Errorf("organization id is required: use -i or --id")
	}
	if offset < 0 {
		return fmt.Errorf("offset is required: use --offset")
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	org, err := cli.GetOrganization(trimmedID)
	if err != nil {
		return err
	}

	if offset < 0 || offset >= len(org.ExternalConnections) {
		return fmt.Errorf("offset %d is out of range: organization %q has %d connection(s)", offset, trimmedID, len(org.ExternalConnections))
	}

	connectionID := org.ExternalConnections[offset].ID
	if _, err := cli.RemoveOrganizationExternalConnection(trimmedID, connectionID); err != nil {
		return err
	}

	fmt.Printf("offset = %d\n", offset)
	return nil
}

func displayExternalConnectionsAsTable(conns []client.ExternalConnection) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Offset", "Type", "Destination"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(conns) {
		table.Append([]string{"", "No external connections available", ""})
		table.Render()
		return
	}

	for i, conn := range conns {
		connCopy := conn
		table.Append([]string{strconv.Itoa(i), conn.Type, connectionSummary(&connCopy)})
	}
	table.Render()
}
