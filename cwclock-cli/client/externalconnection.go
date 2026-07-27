package client

// ExternalConnection is one external storage destination an organization's
// invoices (or an export job's reports) can be pushed to. Only the fields
// relevant to Type are populated: s3 uses Endpoint/BucketName/Region/
// AccessKey/SecretKey, google_drive uses ServiceAccountBase64/FolderID, git
// uses RepoURL plus either Username/Password or SSHPrivateKey/
// SSHPrivateKeyPassphrase.
type ExternalConnection struct {
	ID                      string `json:"id,omitempty"`
	Type                    string `json:"type"`
	Endpoint                string `json:"endpoint,omitempty"`
	BucketName              string `json:"bucketName,omitempty"`
	Region                  string `json:"region,omitempty"`
	AccessKey               string `json:"accessKey,omitempty"`
	SecretKey               string `json:"secretKey,omitempty"`
	ServiceAccountBase64    string `json:"serviceAccountBase64,omitempty"`
	FolderID                string `json:"folderId,omitempty"`
	RepoURL                 string `json:"repoUrl,omitempty"`
	Username                string `json:"username,omitempty"`
	Password                string `json:"password,omitempty"`
	SSHPrivateKey           string `json:"sshPrivateKey,omitempty"`
	SSHPrivateKeyPassphrase string `json:"sshPrivateKeyPassphrase,omitempty"`
	Path                    string `json:"path,omitempty"`
	FlatDirectory           bool   `json:"flatDirectory,omitempty"`
}
