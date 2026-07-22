package externalconn

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

const (
	driveAPIBase       = "https://www.googleapis.com/drive/v3/files"
	driveUploadAPIBase = "https://www.googleapis.com/upload/drive/v3/files"
	driveScope         = "https://www.googleapis.com/auth/drive"
	driveFolderMime    = "application/vnd.google-apps.folder"
	defaultTokenURI    = "https://oauth2.googleapis.com/token"
)

// serviceAccountKey is the subset of a GCP service account JSON key file
// this package needs.
type serviceAccountKey struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// DecodeServiceAccount base64-decodes and parses a service account key,
// exported so the organization handler can validate a connection's
// serviceAccountBase64 field at save time instead of only discovering a
// malformed key the first time an invoice is pushed.
func DecodeServiceAccount(base64JSON string) (email string, err error) {
	key, _, err := parseServiceAccount(base64JSON)
	if err != nil {
		return utils.EMPTY, err
	}
	return key.ClientEmail, nil
}

func parseServiceAccount(base64JSON string) (serviceAccountKey, *rsa.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64JSON))
	if err != nil {
		return serviceAccountKey{}, nil, fmt.Errorf("invalid base64: %w", err)
	}
	var key serviceAccountKey
	if err := json.Unmarshal(raw, &key); err != nil {
		return serviceAccountKey{}, nil, fmt.Errorf("invalid service account JSON: %w", err)
	}

	if utils.IsBlank(key.ClientEmail) || utils.IsBlank(key.PrivateKey) {
		return serviceAccountKey{}, nil, fmt.Errorf("service account JSON is missing client_email or private_key")
	}

	if utils.IsBlank(key.TokenURI) {
		key.TokenURI = defaultTokenURI
	}

	block, _ := pem.Decode([]byte(key.PrivateKey))
	if block == nil {
		return serviceAccountKey{}, nil, fmt.Errorf("service account private_key is not valid PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return serviceAccountKey{}, nil, fmt.Errorf("could not parse service account private key: %w", err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return serviceAccountKey{}, nil, fmt.Errorf("service account private key is not an RSA key")
	}
	return key, rsaKey, nil
}

// driveTarget talks to the Google Drive v3 REST API directly (no
// google-api-go-client dependency - see ai-instruct-39 plan), authenticating
// as a service account via a hand-signed JWT bearer assertion.
type driveTarget struct {
	key        serviceAccountKey
	privateKey *rsa.PrivateKey
	rootFolder string
	flat       bool
	basePath   string
	httpClient *http.Client
}

func newDriveTarget(conn models.ExternalConnection) (*driveTarget, error) {
	key, privateKey, err := parseServiceAccount(conn.ServiceAccountBase64)
	if err != nil {
		return nil, fmt.Errorf("external connection google_drive: %w", err)
	}
	return &driveTarget{
		key:        key,
		privateKey: privateKey,
		rootFolder: conn.FolderID,
		flat:       conn.FlatDirectory,
		basePath:   cleanBasePath(conn.Path),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// pathSegments splits basePath into the individual folder names Drive's
// find-or-create-by-name API needs one call per level for (Drive addresses
// folders by id, not by a path string like S3/git).
func (d *driveTarget) pathSegments() []string {
	if utils.IsBlank(d.basePath) {
		return nil
	}
	return strings.Split(d.basePath, "/")
}

// ensureBaseFolder walks pathSegments from rootFolder, creating whichever
// level is missing, the same optional subfolder git/S3 connections already
// support.
func (d *driveTarget) ensureBaseFolder(ctx context.Context, token string) (string, error) {
	folder := d.rootFolder
	for _, segment := range d.pathSegments() {
		next, err := d.ensureFolder(ctx, token, segment, folder)
		if err != nil {
			return utils.EMPTY, err
		}
		folder = next
	}
	return folder, nil
}

// findBaseFolder is ensureBaseFolder's search-only counterpart, used by
// Delete: stops (not found) as soon as any path segment doesn't exist.
func (d *driveTarget) findBaseFolder(ctx context.Context, token string) (string, bool, error) {
	folder := d.rootFolder
	for _, segment := range d.pathSegments() {
		next, found, err := d.findFolder(ctx, token, segment, folder)
		if err != nil || !found {
			return utils.EMPTY, false, err
		}
		folder = next
	}
	return folder, true, nil
}

func (d *driveTarget) Upload(ctx context.Context, year string, months []string, filename string, data []byte) error {
	token, err := d.accessToken(ctx)
	if err != nil {
		return err
	}

	folder, err := d.ensureTargetFolder(ctx, token, year, months)
	if err != nil {
		return err
	}

	fileID, found, err := d.findFile(ctx, token, filename, folder)
	if err != nil {
		return err
	}
	if found {
		return d.updateFile(ctx, token, fileID, data)
	}
	return d.createFile(ctx, token, filename, folder, data)
}

func (d *driveTarget) Delete(ctx context.Context, year string, months []string, filename string) error {
	token, err := d.accessToken(ctx)
	if err != nil {
		return err
	}

	folder, found, err := d.findTargetFolder(ctx, token, year, months)
	if err != nil || !found {
		return err
	}
	fileID, found, err := d.findFile(ctx, token, filename, folder)
	if err != nil || !found {
		return err
	}
	return d.deleteFile(ctx, token, fileID)
}

// ensureTargetFolder resolves the folder invoices are uploaded into:
// basePath (if set) directly in flat mode (ai-instruct-42, for accounting
// software that needs a flat listing with no subfolders), or basePath's
// year/month folder chain (creating whichever level is missing) otherwise.
func (d *driveTarget) ensureTargetFolder(ctx context.Context, token, year string, months []string) (string, error) {
	base, err := d.ensureBaseFolder(ctx, token)
	if err != nil {
		return utils.EMPTY, err
	}
	if d.flat {
		return base, nil
	}
	yearFolder, err := d.ensureFolder(ctx, token, year, base)
	if err != nil {
		return utils.EMPTY, err
	}
	return d.ensureAnyFolder(ctx, token, months, yearFolder)
}

// findTargetFolder is ensureTargetFolder's search-only counterpart, used by
// Delete: basePath directly in flat mode, or its year/month folder chain if
// every level of it already exists.
func (d *driveTarget) findTargetFolder(ctx context.Context, token, year string, months []string) (string, bool, error) {
	base, found, err := d.findBaseFolder(ctx, token)
	if err != nil || !found {
		return utils.EMPTY, false, err
	}
	if d.flat {
		return base, true, nil
	}
	yearFolder, found, err := d.findFolder(ctx, token, year, base)
	if err != nil || !found {
		return utils.EMPTY, false, err
	}
	return d.findAnyFolder(ctx, token, months, yearFolder)
}

// accessToken exchanges a fresh service-account JWT assertion for a Drive
// access token. Invoice actions are infrequent enough that fetching a new
// token per call isn't worth caching.
func (d *driveTarget) accessToken(ctx context.Context) (string, error) {
	now := time.Now().UTC()
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"iss":   d.key.ClientEmail,
		"scope": driveScope,
		"aud":   d.key.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)
	signingInput := base64URLEncode(headerJSON) + "." + base64URLEncode(claimsJSON)

	hashed := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, d.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: could not sign JWT: %w", err)
	}
	assertion := signingInput + "." + base64URLEncode(signature)

	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.key.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return utils.EMPTY, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: token request failed: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: invalid token response: %w", err)
	}

	if resp.StatusCode >= 300 || utils.IsBlank(tokenResp.AccessToken) {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: token exchange returned %d: %s", resp.StatusCode, tokenResp.Error)
	}

	return tokenResp.AccessToken, nil
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// escapeDriveQueryValue escapes a value embedded in a Drive `q` search
// expression, per Drive's search-query syntax (backslash and single quote
// must be backslash-escaped).
func escapeDriveQueryValue(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}

type driveFileList struct {
	Files []struct {
		ID string `json:"id"`
	} `json:"files"`
}

// searchOne runs a Drive `q` query and returns the first matching file/folder id.
func (d *driveTarget) searchOne(ctx context.Context, token, query string) (id string, found bool, err error) {
	q := url.Values{
		"q":                         {query},
		"fields":                    {"files(id)"},
		"supportsAllDrives":         {"true"},
		"includeItemsFromAllDrives": {"true"},
		"spaces":                    {"drive"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, driveAPIBase+"?"+q.Encode(), nil)
	if err != nil {
		return utils.EMPTY, false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return utils.EMPTY, false, fmt.Errorf("external connection google_drive: search failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return utils.EMPTY, false, fmt.Errorf("external connection google_drive: search returned %d: %s", resp.StatusCode, string(body))
	}

	var list driveFileList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return utils.EMPTY, false, fmt.Errorf("external connection google_drive: invalid search response: %w", err)
	}

	if len(list.Files) == 0 {
		return utils.EMPTY, false, nil
	}

	return list.Files[0].ID, true, nil
}

func (d *driveTarget) findFolder(ctx context.Context, token, name, parentID string) (string, bool, error) {
	query := fmt.Sprintf(
		"name = '%s' and '%s' in parents and mimeType = '%s' and trashed = false",
		escapeDriveQueryValue(name), parentID, driveFolderMime,
	)
	return d.searchOne(ctx, token, query)
}

func (d *driveTarget) ensureFolder(ctx context.Context, token, name, parentID string) (string, error) {
	if id, found, err := d.findFolder(ctx, token, name, parentID); err != nil {
		return utils.EMPTY, err
	} else if found {
		return id, nil
	}

	return d.createFolder(ctx, token, name, parentID)
}

// findAnyFolder searches for a folder matching any of names (e.g. the
// month's English and French candidate names), so an existing folder in
// either language is found rather than only the default one.
func (d *driveTarget) findAnyFolder(ctx context.Context, token string, names []string, parentID string) (string, bool, error) {
	clauses := make([]string, len(names))
	for i, name := range names {
		clauses[i] = fmt.Sprintf("name = '%s'", escapeDriveQueryValue(name))
	}

	query := fmt.Sprintf(
		"(%s) and '%s' in parents and mimeType = '%s' and trashed = false",
		strings.Join(clauses, " or "), parentID, driveFolderMime,
	)

	return d.searchOne(ctx, token, query)
}

// ensureAnyFolder reuses whichever of names' candidate folders already
// exists, creating the first (default) candidate if none do.
func (d *driveTarget) ensureAnyFolder(ctx context.Context, token string, names []string, parentID string) (string, error) {
	if id, found, err := d.findAnyFolder(ctx, token, names, parentID); err != nil {
		return utils.EMPTY, err
	} else if found {
		return id, nil
	}

	return d.createFolder(ctx, token, names[0], parentID)
}

func (d *driveTarget) createFolder(ctx context.Context, token, name, parentID string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"name":     name,
		"mimeType": driveFolderMime,
		"parents":  []string{parentID},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, driveAPIBase+"?fields=id&supportsAllDrives=true", bytes.NewReader(body))
	if err != nil {
		return utils.EMPTY, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: create folder failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return utils.EMPTY, fmt.Errorf("external connection google_drive: create folder returned %d: %s", resp.StatusCode, string(respBody))
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return utils.EMPTY, fmt.Errorf("external connection google_drive: invalid create folder response: %w", err)
	}
	return created.ID, nil
}

func (d *driveTarget) findFile(ctx context.Context, token, name, parentID string) (string, bool, error) {
	query := fmt.Sprintf(
		"name = '%s' and '%s' in parents and mimeType != '%s' and trashed = false",
		escapeDriveQueryValue(name), parentID, driveFolderMime,
	)
	return d.searchOne(ctx, token, query)
}

func (d *driveTarget) createFile(ctx context.Context, token, name, parentID string, data []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadata, _ := json.Marshal(map[string]any{"name": name, "parents": []string{parentID}})
	metaPart, err := writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json; charset=UTF-8"}})
	if err != nil {
		return err
	}
	if _, err := metaPart.Write(metadata); err != nil {
		return err
	}

	mediaPart, err := writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/pdf"}})
	if err != nil {
		return err
	}
	if _, err := mediaPart.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, driveUploadAPIBase+"?uploadType=multipart&supportsAllDrives=true", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/related; boundary="+writer.Boundary())

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("external connection google_drive: upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("external connection google_drive: upload returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (d *driveTarget) updateFile(ctx context.Context, token, fileID string, data []byte) error {
	reqURL := driveUploadAPIBase + "/" + url.PathEscape(fileID) + "?uploadType=media&supportsAllDrives=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/pdf")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("external connection google_drive: replace failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("external connection google_drive: replace returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (d *driveTarget) deleteFile(ctx context.Context, token, fileID string) error {
	reqURL := driveAPIBase + "/" + url.PathEscape(fileID) + "?supportsAllDrives=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("external connection google_drive: delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("external connection google_drive: delete returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
