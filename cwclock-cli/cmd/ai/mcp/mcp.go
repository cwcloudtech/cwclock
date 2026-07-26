package mcp

import (
	"cwclock/utils"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	mcp_http_transport "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/spf13/cobra"
)

var (
	port       int
	endpoint   string
	listenAddr string
)

//go:embed instructions.md
var embeddedInstructions string

type runCwclockCommandArgs struct {
	Command string   `json:"command" jsonschema:"required,description=The cwclock command to run without the leading cwclock binary name"`
	Args    []string `json:"args" jsonschema:"description=Additional command arguments and flags"`
}

type getCwclockCommandHelpArgs struct {
	Command string `json:"command" jsonschema:"required,description=Top-level cwclock command to get help for (e.g. configure, bootstrap, ai)"`
}

type dynamicCommandArgs struct {
	Args []string `json:"args" jsonschema:"description=Additional args, subcommands and flags for this command path. For usage examples and exact flag mappings, call get_mcp_usage_guide."`
}

type emptyToolArgs struct{}

type commandSpec struct {
	Path        []string
	Description string
}

// McpCmd represents the MCP command group under ai.
var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the cwclock MCP server",
	Long:  "Start the cwclock MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%d", listenAddr, port)
		transport := mcp_http_transport.NewHTTPTransport(endpoint).WithAddr(addr)

		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to resolve cwclock executable: %w", err)
		}

		dynamicCommands, err := discoverCwclockCommands(executable)
		if err != nil {
			return fmt.Errorf("failed to discover cwclock commands: %w", err)
		}

		dynamicToolLines := make([]string, 0, len(dynamicCommands))
		for _, spec := range dynamicCommands {
			if len(spec.Path) == 0 {
				continue
			}
			if len(spec.Path) >= 2 && spec.Path[0] == "ai" && spec.Path[1] == "mcp" {
				continue
			}
			toolName := "cwclock_" + sanitizeToolName(strings.Join(spec.Path, "_"))
			desc := strings.TrimSpace(spec.Description)
			if utils.IsBlank(desc) {
				desc = "(no description)"
			}

			dynamicToolLines = append(dynamicToolLines, fmt.Sprintf("- %s => cwclock %s | %s", toolName, strings.Join(spec.Path, " "), desc))
		}
		sort.Strings(dynamicToolLines)

		server := mcp_golang.NewServer(
			transport,
			mcp_golang.WithName("cwclock-mcp-server"),
			mcp_golang.WithVersion("0.1.0"),
			mcp_golang.WithInstructions(strings.TrimSpace(embeddedInstructions)),
		)

		err = server.RegisterTool(
			"get_mcp_usage_guide",
			"Return the embedded cwclock MCP usage guide in Markdown.",
			func(arguments emptyToolArgs) (*mcp_golang.ToolResponse, error) {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(strings.TrimSpace(embeddedInstructions))), nil
			},
		)
		if err != nil {
			return err
		}

		err = server.RegisterTool(
			"list_cwclock_commands",
			"List top-level cwclock commands by returning `cwclock --help` output. Use this first before calling run_cwclock_command.",
			func(arguments emptyToolArgs) (*mcp_golang.ToolResponse, error) {
				runCmd := exec.Command(executable, "--help")
				output, err := runCmd.CombinedOutput()

				exitCode := 0
				if runCmd.ProcessState != nil {
					exitCode = runCmd.ProcessState.ExitCode()
				}

				result := fmt.Sprintf("command: cwclock --help\nexit_code: %d\noutput:\n%s", exitCode, string(output))
				if err != nil {
					return nil, fmt.Errorf("%s", result)
				}

				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
			},
		)
		if err != nil {
			return err
		}

		err = server.RegisterTool(
			"get_cwclock_command_help",
			"Get help for a specific top-level cwclock command by returning `cwclock <command> --help` output.",
			func(arguments getCwclockCommandHelpArgs) (*mcp_golang.ToolResponse, error) {
				commandName := strings.TrimSpace(arguments.Command)
				if utils.IsBlank(commandName) {
					return nil, fmt.Errorf("command is required")
				}

				runCmd := exec.Command(executable, commandName, "--help")
				output, err := runCmd.CombinedOutput()

				exitCode := 0
				if runCmd.ProcessState != nil {
					exitCode = runCmd.ProcessState.ExitCode()
				}

				result := fmt.Sprintf("command: cwclock %s --help\nexit_code: %d\noutput:\n%s", commandName, exitCode, string(output))
				if err != nil {
					return nil, fmt.Errorf("%s", result)
				}

				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
			},
		)
		if err != nil {
			return err
		}

		err = server.RegisterTool(
			"list_mcp_dynamic_tools",
			"List all dynamically generated MCP tools and their mapped cwclock command paths.",
			func(arguments emptyToolArgs) (*mcp_golang.ToolResponse, error) {
				if len(dynamicToolLines) == 0 {
					return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No dynamic tools discovered.")), nil
				}
				result := fmt.Sprintf("discovered_dynamic_tools: %d\n%s", len(dynamicToolLines), strings.Join(dynamicToolLines, "\n"))
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
			},
		)
		if err != nil {
			return err
		}

		for _, spec := range dynamicCommands {
			if len(spec.Path) == 0 {
				continue
			}

			if len(spec.Path) >= 2 && spec.Path[0] == "ai" && spec.Path[1] == "mcp" {
				continue
			}

			toolName := "cwclock_" + sanitizeToolName(strings.Join(spec.Path, "_"))
			desc := strings.TrimSpace(spec.Description)
			if utils.IsBlank(desc) {
				desc = fmt.Sprintf("Run command path: cwclock %s", strings.Join(spec.Path, " "))
			}

			fullPath := append([]string{}, spec.Path...)

			err = server.RegisterTool(
				toolName,
				fmt.Sprintf("%s. Base command: cwclock %s", desc, strings.Join(fullPath, " ")),
				func(arguments dynamicCommandArgs) (*mcp_golang.ToolResponse, error) {
					commandName := normalizeCommandName(fullPath[0])
					cliArgs := append([]string{}, fullPath...)
					cliArgs[0] = commandName
					cliArgs = append(cliArgs, arguments.Args...)
					result, runErr := runCLI(executable, cliArgs)
					if runErr != nil {
						return nil, fmt.Errorf("%s", result)
					}
					return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
				},
			)
			if err != nil {
				return err
			}
		}

		err = server.RegisterTool(
			"run_cwclock_command",
			"Fallback generic command runner. Prefer dynamic cwclock_<command_path> tools discovered from Cobra tree.",
			func(arguments runCwclockCommandArgs) (*mcp_golang.ToolResponse, error) {
				commandName := strings.TrimSpace(arguments.Command)
				cliArgs := make([]string, 0, 1+len(arguments.Args))
				if utils.IsNotBlank(commandName) {
					cliArgs = append(cliArgs, strings.Fields(commandName)...)
				}

				cliArgs = append(cliArgs, arguments.Args...)

				for len(cliArgs) > 0 && strings.EqualFold(strings.TrimSpace(cliArgs[0]), "cwclock") {
					cliArgs = cliArgs[1:]
				}

				if len(cliArgs) == 0 {
					return nil, fmt.Errorf("command is required")
				}

				commandName = strings.TrimSpace(cliArgs[0])
				commandArgs := make([]string, 0, len(cliArgs)-1)
				if len(cliArgs) > 1 {
					commandArgs = append(commandArgs, cliArgs[1:]...)
				}

				commandName = normalizeCommandName(commandName)

				if strings.HasSuffix(strings.ToLower(commandName), "s") && len(commandName) > 1 {
					commandName = strings.TrimSuffix(commandName, "s")
				}

				if commandName == "ai" && len(arguments.Args) > 0 && arguments.Args[0] == "mcp" {
					return nil, fmt.Errorf("running ai mcp from the MCP tool is blocked")
				}

				cliArgs = append([]string{commandName}, commandArgs...)

				result, err := runCLI(executable, cliArgs)
				if err != nil {
					return nil, fmt.Errorf("%s", result)
				}

				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
			},
		)
		if err != nil {
			return err
		}

		fmt.Printf("Starting cwclock MCP server on http://%s%s\n", addr, endpoint)
		return server.Serve()
	},
}

func init() {
	McpCmd.DisableFlagsInUseLine = true
	McpCmd.Flags().StringVarP(&listenAddr, "listen", "l", "127.0.0.1", "MCP server listen address")
	McpCmd.Flags().IntVarP(&port, "port", "p", 8080, "MCP server port")
	McpCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "/mcp", "MCP HTTP endpoint path")
}

func runCLI(executable string, cliArgs []string) (string, error) {
	runCmd := exec.Command(executable, cliArgs...)
	output, err := runCmd.CombinedOutput()

	exitCode := 0
	if runCmd.ProcessState != nil {
		exitCode = runCmd.ProcessState.ExitCode()
	}

	result := fmt.Sprintf("command: cwclock %s\nexit_code: %d\noutput:\n%s", strings.Join(cliArgs, " "), exitCode, string(output))
	if err != nil {
		return result, err
	}
	return result, nil
}

func discoverCwclockCommands(executable string) ([]commandSpec, error) {
	type queueEntry struct {
		Path []string
	}

	queue := []queueEntry{{Path: []string{}}}
	visited := map[string]bool{"": true}
	collected := map[string]commandSpec{}

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		helpArgs := append([]string{}, entry.Path...)
		helpArgs = append(helpArgs, "--help")
		runCmd := exec.Command(executable, helpArgs...)
		output, err := runCmd.CombinedOutput()
		if err != nil {
			continue
		}

		subCommands := parseAvailableCommands(string(output))
		for _, sub := range subCommands {
			path := append(append([]string{}, entry.Path...), sub.Name)
			key := strings.Join(path, " ")
			if !visited[key] {
				visited[key] = true
				queue = append(queue, queueEntry{Path: path})
			}
			collected[key] = commandSpec{Path: path, Description: sub.Description}
		}
	}

	result := make([]commandSpec, 0, len(collected))
	keys := make([]string, 0, len(collected))
	for key := range collected {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, collected[key])
	}

	return result, nil
}

type parsedCommand struct {
	Name        string
	Description string
}

func parseAvailableCommands(helpText string) []parsedCommand {
	lines := strings.Split(helpText, "\n")
	commands := make([]parsedCommand, 0)
	inAvailable := false
	re := regexp.MustCompile(`^\s{2,}([a-zA-Z0-9_-]+)\s{2,}(.*)$`)

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !inAvailable {
			if strings.HasPrefix(trim, "Available Commands:") {
				inAvailable = true
			}
			continue
		}

		if utils.IsBlank(trim) {
			continue
		}

		if strings.HasSuffix(trim, ":") {
			break
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		name := strings.TrimSpace(matches[1])
		if name == "help" {
			continue
		}

		commands = append(commands, parsedCommand{Name: name, Description: strings.TrimSpace(matches[2])})
	}

	return commands
}

func sanitizeToolName(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "-", "_"), " ", "_")

	b := strings.Builder{}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if utils.IsBlank(out) {
		return "command"
	}

	return out
}

func normalizeCommandName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
