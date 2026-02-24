package cmd

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/internal/buildinfo"
	mcppkg "icekalt.dev/money-tracker/internal/mcp"

	"github.com/spf13/cobra"
)

var (
	mcpURL   string
	mcpToken string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP (Model Context Protocol) server over stdio",
	Long: `Start an MCP server that communicates over stdio (JSON-RPC).
This allows LLM clients (Claude Desktop, Claude Code, etc.) to interact
with Money Tracker — create transactions, query summaries, manage households.

The MCP server connects to a running Money Tracker API server via HTTP.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Flags take precedence, then fall back to config (which binds env vars)
		url := mcpURL
		if !cmd.Flags().Changed("url") && cfg.MCP.URL != "" {
			url = cfg.MCP.URL
		}
		token := mcpToken
		if !cmd.Flags().Changed("token") && cfg.MCP.Token != "" {
			token = cfg.MCP.Token
		}

		if token == "" {
			return fmt.Errorf("API token is required (use --token or MONEY_TRACKER_MCP_TOKEN env var)")
		}

		client := mcppkg.NewClient(url, token)
		server := mcppkg.NewServer(client, buildinfo.Version)

		return server.Run(context.Background())
	},
}

func init() {
	mcpCmd.Flags().StringVar(&mcpURL, "url", "http://localhost:8080", "Money Tracker API base URL")
	mcpCmd.Flags().StringVar(&mcpToken, "token", "", "API token for authentication")
	rootCmd.AddCommand(mcpCmd)
}
