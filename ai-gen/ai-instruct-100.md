# AI instruction 100

## Cli

I want this method:

```go
func resolveProject(orgID string, identifier string) (client.Project, error) {
	trimmed := strings.TrimSpace(identifier)

	cli, err := client.NewClient()
	if err != nil {
		return client.Project{}, err
	}
	projects, err := cli.ListProjects(orgID, utils.EMPTY)
	if err != nil {
		return client.Project{}, err
	}

	if looksLikeUUID(trimmed) {
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
```

To path the clientId because most of the time it's known.

And here:

```go
func (c *Client) ListProjects(orgID string, clientID string) ([]Project, error) {
	path := projectsPath(orgID)
	if utils.IsNotBlank(clientID) {
		path += "?clientId=" + url.QueryEscape(clientID)
	}
```

Replace `utils.IsNotBlank` by `utils.IsValidUUID`.
Refactor every other place where you test the validatity of a uuid by this method.
