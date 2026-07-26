package connection

import (
	"cwclock/cmd/organization/connection/ls"
	"cwclock/handlers"
	"cwclock/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	id     string
	format string

	connType                string
	endpoint                string
	bucketName              string
	region                  string
	accessKey               string
	secretKey               string
	serviceAccountBase64    string
	folderID                string
	repoURL                 string
	username                string
	password                string
	sshPrivateKey           string
	sshPrivateKeyPassphrase string
	path                    string
	flatDirectory           bool

	offset int
)

var ConnectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Manage an organization's external connections",
	Long:  `This command lets you list, create and delete an organization's external storage connections (s3, google_drive, git).`,
	Run: func(cmd *cobra.Command, args []string) {
		hasOffset := cmd.Flags().Changed("offset")
		hasType := cmd.Flags().Changed("type")

		if hasOffset && hasType {
			fmt.Println("Error: specify only one of --type (create) or --offset (delete)")
			cmd.Help()
			utils.ExitIfNeeded("", true)
			return
		}

		var err error
		switch {
		case hasOffset:
			err = handlers.HandleOrgConnectionDelete(id, offset)
		case hasType:
			fields := handlers.OrgConnectionFields{
				Type: connType, Endpoint: endpoint, BucketName: bucketName, Region: region,
				AccessKey: accessKey, SecretKey: secretKey, ServiceAccountBase64: serviceAccountBase64,
				FolderID: folderID, RepoURL: repoURL, Username: username, Password: password,
				SSHPrivateKey: sshPrivateKey, SSHPrivateKeyPassphrase: sshPrivateKeyPassphrase,
				Path: path, FlatDirectory: flatDirectory,
			}
			err = handlers.HandleOrgConnectionCreate(id, fields, format)
		default:
			cmd.Help()
			return
		}

		utils.ExitIfError(err)
	},
}

func init() {
	ConnectionCmd.DisableFlagsInUseLine = true
	ConnectionCmd.Flags().StringVarP(&id, "id", "i", "", "Organization ID (required)")
	ConnectionCmd.Flags().StringVar(&connType, "type", "", "Connection type: s3, google_drive or git (create)")
	ConnectionCmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint (s3)")
	ConnectionCmd.Flags().StringVar(&bucketName, "bucket-name", "", "S3 bucket name (s3)")
	ConnectionCmd.Flags().StringVar(&region, "region", "", "S3 region (s3)")
	ConnectionCmd.Flags().StringVar(&accessKey, "access-key", "", "S3 access key (s3)")
	ConnectionCmd.Flags().StringVar(&secretKey, "secret-key", "", "S3 secret key (s3)")
	ConnectionCmd.Flags().StringVar(&serviceAccountBase64, "service-account-base64", "", "Base64-encoded Google service account (google_drive)")
	ConnectionCmd.Flags().StringVar(&folderID, "folder-id", "", "Google Drive folder ID (google_drive)")
	ConnectionCmd.Flags().StringVar(&repoURL, "repo-url", "", "Git repository URL (git)")
	ConnectionCmd.Flags().StringVar(&username, "username", "", "Git username (git, with --password)")
	ConnectionCmd.Flags().StringVar(&password, "password", "", "Git password (git, with --username)")
	ConnectionCmd.Flags().StringVar(&sshPrivateKey, "ssh-private-key", "", "Git SSH private key (git)")
	ConnectionCmd.Flags().StringVar(&sshPrivateKeyPassphrase, "ssh-private-key-passphrase", "", "Git SSH private key passphrase (git)")
	ConnectionCmd.Flags().StringVar(&path, "path", "", "Optional destination subfolder")
	ConnectionCmd.Flags().BoolVar(&flatDirectory, "flat-directory", false, "Upload directly at the destination root instead of nesting under YYYY/MM")
	ConnectionCmd.Flags().IntVar(&offset, "offset", -1, "Offset of the connection to delete in the connections array (delete)")
	ConnectionCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")

	ConnectionCmd.AddCommand(ls.LsCmd)
}
