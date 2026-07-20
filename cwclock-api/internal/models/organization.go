package models

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleReader Role = "reader"
)

type Organization struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	// AccountingEmail is an optional second address (the organization's
	// accounting department, when it has one distinct from the owner) that
	// always gets a copy of invoice emails sent to a client, alongside the
	// owner.
	AccountingEmail      string `json:"accountingEmail,omitempty"`
	Address              string `json:"address"`
	PostalCode           string `json:"postalCode"`
	City                 string `json:"city"`
	Country              string `json:"country"`
	VATNumber            string `json:"vatNumber"`
	SIREN                string `json:"siren"`
	SIRET                string `json:"siret"`
	NAF                  string `json:"naf"`
	MF                   string `json:"mf"`
	IdentificationNumber string `json:"identificationNumber"`
	// IBAN/BIC are optional bank details for wire-transfer payment. Both
	// must be set for either to appear on an invoice PDF (see
	// report.formatIBANBIC) - either alone isn't enough to actually pay by
	// bank transfer.
	IBAN                string               `json:"iban,omitempty"`
	BIC                 string               `json:"bic,omitempty"`
	Picture             string               `json:"picture,omitempty"`
	PictureX            float64              `json:"pictureX"`
	PictureY            float64              `json:"pictureY"`
	Stamp               string               `json:"stamp,omitempty"`
	StampX              float64              `json:"stampX"`
	StampY              float64              `json:"stampY"`
	Currency            string               `json:"currency"`
	ExternalConnections []ExternalConnection `json:"externalConnections"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

// ExternalConnectionType is the kind of external storage an organization can
// push generated invoices to.
type ExternalConnectionType string

const (
	ExternalConnectionS3          ExternalConnectionType = "s3"
	ExternalConnectionGoogleDrive ExternalConnectionType = "google_drive"
)

// ExternalConnection is one optional external storage destination an
// organization's invoices get pushed to (see ai-instruct-39). Only the
// fields relevant to Type are populated - S3 connections use
// Endpoint/BucketName/Region/AccessKey/SecretKey, Google Drive connections
// use ServiceAccountBase64/FolderID.
type ExternalConnection struct {
	ID                   string                 `json:"id"`
	Type                 ExternalConnectionType `json:"type"`
	Endpoint             string                 `json:"endpoint,omitempty"`
	BucketName           string                 `json:"bucketName,omitempty"`
	Region               string                 `json:"region,omitempty"`
	AccessKey            string                 `json:"accessKey,omitempty"`
	SecretKey            string                 `json:"secretKey,omitempty"`
	ServiceAccountBase64 string                 `json:"serviceAccountBase64,omitempty"`
	FolderID             string                 `json:"folderId,omitempty"`
	// FlatDirectory, when true, uploads invoices directly at the
	// destination's root instead of nesting them under a "YYYY/MM.MonthName"
	// folder (ai-instruct-42) - some accounting software watching a Drive
	// folder or S3 bucket requires a flat listing with no subfolders. False
	// (the zero value, so existing connections keep today's behavior) nests
	// by year/month as before.
	FlatDirectory bool `json:"flatDirectory,omitempty"`
}

// OrganizationWithOwner adds the owner's email to an Organization, for the
// superuser's organization-management screen (which lists orgs the caller
// isn't necessarily a member of, so it can't resolve the owner client-side).
type OrganizationWithOwner struct {
	Organization
	OwnerEmail string `json:"ownerEmail"`
}

type Member struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Surname        string    `json:"surname"`
	Role           Role      `json:"role"`
	DailyRate      *float64  `json:"dailyRate,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
