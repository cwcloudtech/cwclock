package externalconn

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

const s3Service = "s3"

// s3Target talks to an S3-compatible object store using hand-signed AWS
// SigV4 requests (no AWS SDK dependency - see ai-instruct-39 plan), so any
// endpoint (AWS itself, MinIO, DigitalOcean Spaces, ...) works as long as it
// speaks the S3 REST API. Path-style addressing (endpoint/bucket/key) is
// used rather than virtual-hosted style, since it works uniformly across
// providers without requiring DNS/TLS for a per-bucket subdomain.
type s3Target struct {
	endpoint   string
	bucket     string
	region     string
	accessKey  string
	secretKey  string
	flat       bool
	basePath   string
	httpClient *http.Client
}

func newS3Target(conn models.ExternalConnection) *s3Target {
	return &s3Target{
		endpoint:   utils.GetBaseUrl(conn.Endpoint),
		bucket:     conn.BucketName,
		region:     conn.Region,
		accessKey:  conn.AccessKey,
		secretKey:  conn.SecretKey,
		flat:       conn.FlatDirectory,
		basePath:   cleanBasePath(conn.Path),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *s3Target) Upload(ctx context.Context, year string, months []string, filename string, data []byte) error {
	key, err := s.resolveKey(ctx, year, months, filename)
	if err != nil {
		return err
	}
	return s.do(ctx, http.MethodPut, key, data, "application/pdf")
}

func (s *s3Target) Delete(ctx context.Context, year string, months []string, filename string) error {
	key, err := s.resolveKey(ctx, year, months, filename)
	if err != nil {
		return err
	}
	return s.do(ctx, http.MethodDelete, key, nil, utils.EMPTY)
}

// resolveKey returns the key of an object already sitting under one of
// months' candidate month folders (so a file previously filed under an
// alternate-language month folder gets replaced/deleted in place rather
// than duplicated), falling back to the first (default) candidate's key if
// none of them currently hold the file. In flat mode (ai-instruct-42) the
// key is always just the filename at the bucket's root (or basePath, if
// set), with no year/month lookup at all. basePath, when set, prefixes
// every key, the same optional subfolder git connections already support.
func (s *s3Target) resolveKey(ctx context.Context, year string, months []string, filename string) (string, error) {
	prefix := utils.EMPTY
	if utils.IsNotBlank(s.basePath) {
		prefix = s.basePath + "/"
	}

	if s.flat {
		return prefix + filename, nil
	}
	for _, month := range months {
		key := prefix + year + "/" + month + "/" + filename
		found, err := s.exists(ctx, key)
		if err != nil {
			return utils.EMPTY, err
		}

		if found {
			return key, nil
		}
	}
	return prefix + year + "/" + months[0] + "/" + filename, nil
}

// exists reports whether key is already present in the bucket, via a signed
// HEAD request.
func (s *s3Target) exists(ctx context.Context, key string) (bool, error) {
	reqURL := s.endpoint + "/" + uriEncodePath(s.bucket) + "/" + uriEncodePath(key)
	u, err := url.Parse(reqURL)
	if err != nil {
		return false, fmt.Errorf("external connection s3: invalid endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		return false, err
	}
	s.sign(req, u, nil)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("external connection s3: HEAD request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("external connection s3: HEAD %s returned %d", key, resp.StatusCode)
	}
}

func (s *s3Target) do(ctx context.Context, method, key string, body []byte, contentType string) error {
	reqURL := s.endpoint + "/" + uriEncodePath(s.bucket) + "/" + uriEncodePath(key)
	u, err := url.Parse(reqURL)
	if err != nil {
		return fmt.Errorf("external connection s3: invalid endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if utils.IsNotBlank(contentType) {
		req.Header.Set("Content-Type", contentType)
	}
	s.sign(req, u, body)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("external connection s3: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("external connection s3: %s %s returned %d: %s", method, key, resp.StatusCode, string(respBody))
	}
	return nil
}

// sign attaches the Host, x-amz-date, x-amz-content-sha256 and Authorization
// headers implementing AWS Signature Version 4 for a single request whose
// entire body is already known (so the payload hash is always the real
// hash, never "UNSIGNED-PAYLOAD").
func (s *s3Target) sign(req *http.Request, u *url.URL, body []byte) {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	payloadHash := hexSHA256(body)

	req.Header.Set("Host", u.Host)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	signedHeaderNames := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if utils.IsNotBlank(req.Header.Get("Content-Type")) {
		signedHeaderNames = append(signedHeaderNames, "content-type")
	}
	sort.Strings(signedHeaderNames)

	var canonicalHeaders strings.Builder
	for _, name := range signedHeaderNames {
		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteByte(':')
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(name)))
		canonicalHeaders.WriteByte('\n')
	}
	signedHeaders := strings.Join(signedHeaderNames, ";")

	canonicalRequest := strings.Join([]string{
		req.Method,
		u.EscapedPath(),
		"",
		canonicalHeaders.String(),
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s3Service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hexSHA256([]byte(canonicalRequest)),
	}, "\n")

	signingKey := s3SigningKey(s.secretKey, dateStamp, s.region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature,
	))
}

func s3SigningKey(secretKey, dateStamp, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, s3Service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func hexSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// uriEncodePath percent-encodes a single path segment per SigV4's rules
// (unreserved characters A-Z a-z 0-9 - _ . ~ pass through, everything else
// is percent-encoded), used for both the bucket name and the object key -
// the key's own "/" separators are preserved by the caller joining
// per-segment encodings rather than encoding the whole key at once.
func uriEncodePath(s string) string {
	segments := strings.Split(s, "/")
	for i, seg := range segments {
		segments[i] = uriEncodeSegment(seg)
	}
	return strings.Join(segments, "/")
}

func uriEncodeSegment(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}
