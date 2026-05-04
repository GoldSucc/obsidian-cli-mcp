package main

import (
	"context"
	"log"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "obsidian-cli-mcp",
		Version: "0.2.2",
	}, nil)

	tools.RegisterGeneric(server)
	tools.RegisterFiles(server)
	tools.RegisterFolders(server)
	tools.RegisterDaily(server)
	tools.RegisterSearch(server)
	tools.RegisterTasks(server)
	tools.RegisterProperties(server)
	tools.RegisterTagsLinks(server)
	tools.RegisterRandom(server)
	tools.RegisterTemplates(server)
	tools.RegisterBases(server)
	tools.RegisterBookmarks(server)
	tools.RegisterHistory(server)
	tools.RegisterSync(server)
	tools.RegisterVault(server)
	tools.RegisterWorkspace(server)
	tools.RegisterCommands(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
