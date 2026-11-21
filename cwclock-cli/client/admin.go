package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// adminListPageSize is the page size the CLI requests when walking the paged
// admin list endpoints to collect every row. The API caps maxResults well
// above this, so each request returns a full page until the last one.
const adminListPageSize = 100

// adminUsersPage / adminOrgsPage mirror cwclock-api's SearchResult envelope
// (startIndex is a 0-based page number, not a row offset).
type adminUsersPage struct {
	TotalResults int64       `json:"totalResults"`
	Results      []AdminUser `json:"results"`
}

type adminOrgsPage struct {
	TotalResults int64               `json:"totalResults"`
	Results      []AdminOrganization `json:"results"`
}

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

// pagedQuery builds a ?startIndex=&maxResults= query string for one page of a
// paged admin list endpoint.
func pagedQuery(startIndex int) string {
	values := url.Values{}
	values.Set("startIndex", strconv.Itoa(startIndex))
	values.Set("maxResults", strconv.Itoa(adminListPageSize))
	return "?" + values.Encode()
}

// AdminListOrganizations walks every page of the paged /admin/organizations
// endpoint and returns the full list, so the CLI keeps its "list all" behaviour
// over the server-paginated response.
func (c *Client) AdminListOrganizations() ([]AdminOrganization, error) {
	var all []AdminOrganization
	for startIndex := 0; ; startIndex++ {
		responseBody, err := c.httpRequest(adminOrganizationsPath()+pagedQuery(startIndex), "GET", bytes.Buffer{})
		if err != nil {
			return nil, err
		}
		var page adminOrgsPage
		decodeErr := json.NewDecoder(responseBody).Decode(&page)
		responseBody.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode admin organizations response: %w", decodeErr)
		}
		all = append(all, page.Results...)
		if len(page.Results) < adminListPageSize || int64(len(all)) >= page.TotalResults {
			break
		}
	}
	return all, nil
}

// AdminListUsers walks every page of the paged /admin/users endpoint and
// returns the full list, so the CLI keeps its "list all" behaviour over the
// server-paginated response.
func (c *Client) AdminListUsers() ([]AdminUser, error) {
	var all []AdminUser
	for startIndex := 0; ; startIndex++ {
		responseBody, err := c.httpRequest(adminUsersPath()+pagedQuery(startIndex), "GET", bytes.Buffer{})
		if err != nil {
			return nil, err
		}
		var page adminUsersPage
		decodeErr := json.NewDecoder(responseBody).Decode(&page)
		responseBody.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode admin users response: %w", decodeErr)
		}
		all = append(all, page.Results...)
		if len(page.Results) < adminListPageSize || int64(len(all)) >= page.TotalResults {
			break
		}
	}
	return all, nil
}

// AdminFindUsers resolves term (an id or an email, or any name fragment)
// against GET /admin/users?term=..., walking every page. The server matches it
// case-insensitively against name/surname/email/id, so this is the CLI's
// id-or-email lookup for `cwclock admin user` (see ai-instruct-99).
func (c *Client) AdminFindUsers(term string) ([]AdminUser, error) {
	var all []AdminUser
	for startIndex := 0; ; startIndex++ {
		values := url.Values{}
		values.Set("term", term)
		values.Set("startIndex", strconv.Itoa(startIndex))
		values.Set("maxResults", strconv.Itoa(adminListPageSize))

		responseBody, err := c.httpRequest(adminUsersPath()+"?"+values.Encode(), "GET", bytes.Buffer{})
		if err != nil {
			return nil, err
		}
		var page adminUsersPage
		decodeErr := json.NewDecoder(responseBody).Decode(&page)
		responseBody.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode admin users response: %w", decodeErr)
		}
		all = append(all, page.Results...)
		if len(page.Results) < adminListPageSize || int64(len(all)) >= page.TotalResults {
			break
		}
	}
	return all, nil
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
