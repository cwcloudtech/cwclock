package handlers

import (
	"cwclock/client"
	"cwclock/utils"
	"fmt"
	"strings"
)

// resolveClient resolves identifier as a client id or, when it isn't
// shaped like one (or isn't found by id), as a client name -
// case-insensitively, picking the first match.
func resolveClient(orgID string, identifier string) (client.OrgClient, error) {
	trimmed := strings.TrimSpace(identifier)

	cli, err := client.NewClient()
	if err != nil {
		return client.OrgClient{}, err
	}
	clients, err := cli.ListOrgClients(orgID)
	if err != nil {
		return client.OrgClient{}, err
	}

	if utils.IsValidUUID(trimmed) {
		for _, c := range clients {
			if c.ID == trimmed {
				return c, nil
			}
		}
	}
	for _, c := range clients {
		if strings.EqualFold(c.Name, trimmed) {
			return c, nil
		}
	}

	return client.OrgClient{}, fmt.Errorf("client %q not found by id or name", trimmed)
}

// resolveClientID is resolveClient's id-only convenience wrapper.
func resolveClientID(orgID string, identifier string) (string, error) {
	c, err := resolveClient(orgID, identifier)
	if err != nil {
		return utils.EMPTY, err
	}
	return c.ID, nil
}

// resolveClientIDs resolves a repeatable --client flag's values (job
// create/update, export); blank input means "every client" and is left
// untouched.
func resolveClientIDs(orgID string, identifiers []string) ([]string, error) {
	normalized := normalizeIDs(identifiers)
	if len(normalized) == 0 {
		return nil, nil
	}
	resolved := make([]string, 0, len(normalized))
	for _, identifier := range normalized {
		id, err := resolveClientID(orgID, identifier)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, id)
	}
	return resolved, nil
}

// resolveProject resolves identifier as a project id or, when it isn't
// shaped like one (or isn't found by id), as a project name -
// case-insensitively, picking the first match. clientID, when known,
// scopes the search to that client's projects instead of the whole
// organization's - most callers already know it by the time they need a
// project, and it also disambiguates two different clients having a
// same-named project. Pass utils.EMPTY when it isn't known.
func resolveProject(orgID string, clientID string, identifier string) (client.Project, error) {
	trimmed := strings.TrimSpace(identifier)

	cli, err := client.NewClient()
	if err != nil {
		return client.Project{}, err
	}
	projects, err := cli.ListProjects(orgID, clientID)
	if err != nil {
		return client.Project{}, err
	}

	if utils.IsValidUUID(trimmed) {
		for _, p := range projects {
			if p.ID == trimmed {
				return p, nil
			}
		}
	}
	for _, p := range projects {
		if strings.EqualFold(p.Name, trimmed) {
			return p, nil
		}
	}

	return client.Project{}, fmt.Errorf("project %q not found by id or name", trimmed)
}

// resolveProjects resolves a repeatable --project flag's values (job
// create/update, invoice, export) to their full project objects; blank
// input means "every project" and is left untouched. Returning the full
// objects (not just ids) lets callers cross-check each project's client
// against a --client scope with requireProjectsMatchClients. clientID is
// resolveProject's same optional scoping hint.
func resolveProjects(orgID string, clientID string, identifiers []string) ([]client.Project, error) {
	normalized := normalizeIDs(identifiers)
	if len(normalized) == 0 {
		return nil, nil
	}
	resolved := make([]client.Project, 0, len(normalized))
	for _, identifier := range normalized {
		p, err := resolveProject(orgID, clientID, identifier)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, p)
	}
	return resolved, nil
}

// resolveProjectIDs is resolveProjects's id-only convenience wrapper.
func resolveProjectIDs(orgID string, clientID string, identifiers []string) ([]string, error) {
	projects, err := resolveProjects(orgID, clientID, identifiers)
	if err != nil {
		return nil, err
	}
	return projectIDsOf(projects), nil
}

// singleClientScope returns clientID when there's exactly one (the search
// can be scoped to it), or utils.EMPTY when there are zero or several
// (ambiguous - search every client's projects instead).
func singleClientScope(clientIDs []string) string {
	if len(clientIDs) == 1 {
		return clientIDs[0]
	}
	return utils.EMPTY
}

// projectIDsOf extracts each project's id, preserving order; nil in, nil
// out.
func projectIDsOf(projects []client.Project) []string {
	if len(projects) == 0 {
		return nil
	}
	ids := make([]string, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	return ids
}

// requireProjectsMatchClients fails fast when both a client and project
// scope were explicitly given and at least one resolved project doesn't
// belong to any of the resolved clients - a project belongs to exactly one
// client, so a mismatch here almost certainly means the user picked an
// inconsistent --client/--project pair by mistake (see ai-instruct-99). A
// blank clientIDs or projects means one side wasn't given, so there's
// nothing to cross-check.
func requireProjectsMatchClients(clientIDs []string, projects []client.Project) error {
	if len(clientIDs) == 0 || len(projects) == 0 {
		return nil
	}

	clientSet := make(map[string]bool, len(clientIDs))
	for _, id := range clientIDs {
		clientSet[id] = true
	}

	for _, p := range projects {
		if !clientSet[p.ClientID] {
			return fmt.Errorf("project %q does not belong to the specified client(s)", p.Name)
		}
	}
	return nil
}

// resolveOrganization resolves identifier as an org id or, when it isn't
// shaped like one (or isn't found by id), as an org name among the orgs the
// current user belongs to - case-insensitively, picking the first match.
func resolveOrganization(identifier string) (client.Organization, error) {
	trimmed := strings.TrimSpace(identifier)

	cli, err := client.NewClient()
	if err != nil {
		return client.Organization{}, err
	}
	orgs, err := cli.ListOrganizations()
	if err != nil {
		return client.Organization{}, err
	}

	if utils.IsValidUUID(trimmed) {
		for _, o := range orgs {
			if o.ID == trimmed {
				return o, nil
			}
		}
	}
	for _, o := range orgs {
		if strings.EqualFold(o.Name, trimmed) {
			return o, nil
		}
	}

	return client.Organization{}, fmt.Errorf("organization %q not found by id or name", trimmed)
}
