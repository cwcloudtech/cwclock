package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// AdminOrganization adds the owner's email to an Organization, mirroring
// cwclock-api's OrganizationWithOwner: the superuser's organization list
// spans orgs they aren't necessarily a member of, so the owner can't be
// resolved client-side the way it is for an organization the caller belongs to.
type AdminOrganization struct {
	Organization
	OwnerEmail string `json:"ownerEmail"`
}

// AdminUser mirrors cwclock-api's UserMeResponse, as returned by the
// superuser-only /admin/users endpoints.
type AdminUser struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	Role       string  `json:"role"`
	Picture    string  `json:"picture,omitempty"`
	PictureX   float64 `json:"pictureX"`
	PictureY   float64 `json:"pictureY"`
	MFAEnabled bool    `json:"mfaEnabled"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// AdminUserPayload mirrors cwclock-api's adminUpdateUserPayload: email,
// name, surname and role are always resent (the API replaces the whole
// resource); Password is only sent when it should be changed.
type AdminUserPayload struct {
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Role     string  `json:"role"`
	Password *string `json:"password,omitempty"`
}

func adminOrganizationsPath() string {
	return "/admin/organizations"
}

func adminUsersPath() string {
	return "/admin/users"
}

func adminUserPath(id string) string {
	return adminUsersPath() + "/" + url.PathEscape(id)
}

func (c *Client) AdminListOrganizations() ([]AdminOrganization, error) {
	responseBody, err := c.httpRequest(adminOrganizationsPath(), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var orgs []AdminOrganization
	if err := json.NewDecoder(responseBody).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("failed to decode admin organizations response: %w", err)
	}
	return orgs, nil
}

func (c *Client) AdminListUsers() ([]AdminUser, error) {
	responseBody, err := c.httpRequest(adminUsersPath(), "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var users []AdminUser
	if err := json.NewDecoder(responseBody).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode admin users response: %w", err)
	}
	return users, nil
}

// AdminFindUsersByEmail resolves one or more emails (OR logic) against
// GET /admin/users?email=..., repeated per email - used as the CLI's
// id-or-email fallback for `cwclock admin user` (see ai-instruct-99).
func (c *Client) AdminFindUsersByEmail(emails []string) ([]AdminUser, error) {
	values := url.Values{}
	for _, email := range emails {
		values.Add("email", email)
	}

	path := adminUsersPath()
	if len(values) > 0 {
		path += "?" + values.Encode()
	}

	responseBody, err := c.httpRequest(path, "GET", bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	var users []AdminUser
	if err := json.NewDecoder(responseBody).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode admin users response: %w", err)
	}
	return users, nil
}

func (c *Client) AdminUpdateUser(id string, payload AdminUserPayload) (AdminUser, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return AdminUser{}, err
	}

	var body bytes.Buffer
	body.Write(encoded)

	responseBody, err := c.httpRequest(adminUserPath(id), "PUT", body)
	if err != nil {
		return AdminUser{}, err
	}
	defer responseBody.Close()

	var user AdminUser
	if err := json.NewDecoder(responseBody).Decode(&user); err != nil {
		return AdminUser{}, fmt.Errorf("failed to decode updated admin user response: %w", err)
	}
	return user, nil
}

func (c *Client) AdminDeleteUser(id string) error {
	responseBody, err := c.httpRequest(adminUserPath(id), "DELETE", bytes.Buffer{})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	_, err = io.ReadAll(responseBody)
	return err
}
