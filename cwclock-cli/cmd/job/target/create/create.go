package create

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID string
	id    string

	targetType              string
	to                      string
	cc                      string
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

	format string
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a target to an export job",
	Long:  `This command lets you add a new target to an existing export job.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetFields := handlers.ExportTargetFields{
			Type: targetType, To: to, CC: cc,
			Endpoint: endpoint, BucketName: bucketName, Region: region,
			AccessKey: accessKey, SecretKey: secretKey,
			ServiceAccountBase64: serviceAccountBase64, FolderID: folderID,
			RepoURL: repoURL, Username: username, Password: password,
			SSHPrivateKey: sshPrivateKey, SSHPrivateKeyPassphrase: sshPrivateKeyPassphrase,
			Path: path,
		}
		err := handlers.HandleJobTargetCreate(orgID, id, targetFields, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	CreateCmd.Flags().StringVarP(&id, "id", "i", "", "Job ID (required)")
	CreateCmd.Flags().StringVar(&targetType, "type", "", "Target type: email (default), s3, google_drive or git")
	CreateCmd.Flags().StringVar(&to, "to", "", "Recipient email(s), comma/semicolon-separated (email target)")
	CreateCmd.Flags().StringVar(&cc, "cc", "", "CC email(s), comma/semicolon-separated (email target)")
	CreateCmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 endpoint (s3 target)")
	CreateCmd.Flags().StringVar(&bucketName, "bucket-name", "", "S3 bucket name (s3 target)")
	CreateCmd.Flags().StringVar(&region, "region", "", "S3 region (s3 target)")
	CreateCmd.Flags().StringVar(&accessKey, "access-key", "", "S3 access key (s3 target)")
	CreateCmd.Flags().StringVar(&secretKey, "secret-key", "", "S3 secret key (s3 target)")
	CreateCmd.Flags().StringVar(&serviceAccountBase64, "service-account-base64", "", "Base64-encoded Google service account (google_drive target)")
	CreateCmd.Flags().StringVar(&folderID, "folder-id", "", "Google Drive folder ID (google_drive target)")
	CreateCmd.Flags().StringVar(&repoURL, "repo-url", "", "Git repository URL (git target)")
	CreateCmd.Flags().StringVar(&username, "username", "", "Git username (git target, with --password)")
	CreateCmd.Flags().StringVar(&password, "password", "", "Git password (git target, with --username)")
	CreateCmd.Flags().StringVar(&sshPrivateKey, "ssh-private-key", "", "Git SSH private key (git target)")
	CreateCmd.Flags().StringVar(&sshPrivateKeyPassphrase, "ssh-private-key-passphrase", "", "Git SSH private key passphrase (git target)")
	CreateCmd.Flags().StringVar(&path, "path", "", "Optional destination subfolder (s3/google_drive/git target)")
	CreateCmd.Flags().StringVarP(&format, "format", "f", "", "Output format override: pretty|json")
}
