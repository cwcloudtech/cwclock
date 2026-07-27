package create

import (
	"cwclock/handlers"
	"cwclock/utils"

	"github.com/spf13/cobra"
)

var (
	orgID            string
	name             string
	cron             string
	reportTypes      string
	timePeriod       string
	clientIDs        string
	projectIDs       string
	includeFinancial bool
	enabled          bool

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
	Short: "Create an export job",
	Long:  `This command lets you create a new scheduled export job, along with its first target.`,
	Run: func(cmd *cobra.Command, args []string) {
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
		targetFields := handlers.ExportTargetFields{
			Type: targetType, To: to, CC: cc,
			Endpoint: endpoint, BucketName: bucketName, Region: region,
			AccessKey: accessKey, SecretKey: secretKey,
			ServiceAccountBase64: serviceAccountBase64, FolderID: folderID,
			RepoURL: repoURL, Username: username, Password: password,
			SSHPrivateKey: sshPrivateKey, SSHPrivateKeyPassphrase: sshPrivateKeyPassphrase,
			Path: path,
		}
		err := handlers.HandleJobCreate(orgID, fields, targetFields, format)
		utils.ExitIfError(err)
	},
}

func init() {
	CreateCmd.DisableFlagsInUseLine = true
	CreateCmd.Flags().StringVarP(&orgID, "org", "o", "", "Organization ID (overrides configured org_id)")
	CreateCmd.Flags().StringVar(&name, "name", "", "Job name (required)")
	CreateCmd.Flags().StringVar(&cron, "cron", "", "Cron expression (required)")
	CreateCmd.Flags().StringVar(&reportTypes, "report-types", "", "Comma-separated report types (required): summary-pdf, summary-csv, detailed-pdf, detailed-csv, unpaid-invoices, all-invoices")
	CreateCmd.Flags().StringVar(&timePeriod, "time-period", "", "Time period covered by each run (required), e.g. now(), now()-1d, now()-1h")
	CreateCmd.Flags().StringVar(&clientIDs, "client-ids", "", "Comma-separated client IDs to include (empty = all)")
	CreateCmd.Flags().StringVar(&projectIDs, "project-ids", "", "Comma-separated project IDs to include (empty = all)")
	CreateCmd.Flags().BoolVar(&includeFinancial, "include-financial", false, "Include financial figures in the exported reports")
	CreateCmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the job is enabled")

	CreateCmd.Flags().StringVar(&targetType, "type", "", "Initial target type: email (default), s3, google_drive or git")
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
