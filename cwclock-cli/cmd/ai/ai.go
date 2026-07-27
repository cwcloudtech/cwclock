package ai

import (
	"cwclock/cmd/ai/agent"
	"cwclock/cmd/ai/mcp"
	"cwclock/cmd/ai/web_agent"

	"github.com/spf13/cobra"
)

var AiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI features and tools",
	Long:  `This command lets you call the AI agent, manage prompts, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	AiCmd.DisableFlagsInUseLine = true
	AiCmd.AddCommand(agent.AgentCmd)
	AiCmd.AddCommand(mcp.McpCmd)
	AiCmd.AddCommand(web_agent.WebAgentCmd)
}
