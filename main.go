package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/google/go-github/v57/github"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "golang.org/x/oauth2"
)

var githubClient *github.Client

func init() {
    // Initialize GitHub client with token
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        log.Fatal("GITHUB_TOKEN environment variable is required")
    }

    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    githubClient = github.NewClient(tc)
}

func main() {
    // Create MCP server
    s := server.NewMCPServer(
        "GitHub Agentic Chat MCP",
        "1.0.0",
        server.WithResourceCapabilities(true, true),
        server.WithLogging(),
    )

    // Add search repositories tool
    searchReposTool := mcp.NewTool("search_repositories",
        mcp.WithDescription("Search GitHub repositories"),
        mcp.WithString("query",
            mcp.Required(),
            mcp.Description("Search query for repositories"),
        ),
    )
    s.AddTool(searchReposTool, handleSearchRepositories)

    // Add create issue tool
    createIssueTool := mcp.NewTool("create_issue",
        mcp.WithDescription("Create a new issue in a repository"),
        mcp.WithString("owner",
            mcp.Required(),
            mcp.Description("Repository owner"),
        ),
        mcp.WithString("repo",
            mcp.Required(),
            mcp.Description("Repository name"),
        ),
        mcp.WithString("title",
            mcp.Required(),
            mcp.Description("Issue title"),
        ),
        mcp.WithString("body",
            mcp.Required(),
            mcp.Description("Issue body"),
        ),
    )
    s.AddTool(createIssueTool, handleCreateIssue)

    // Start the server
    if err := server.ServeStdio(s); err != nil {
        log.Fatalf("Server error: %v\n", err)
    }
}

func handleSearchRepositories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    query := request.Params.Arguments["query"].(string)
    
    opts := &github.SearchOptions{
        ListOptions: github.ListOptions{
            PerPage: 10,
        },
    }
    
    result, _, err := githubClient.Search.Repositories(ctx, query, opts)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Error searching repositories: %v", err)), nil
    }

    var repoList []string
    for _, repo := range result.Repositories {
        repoList = append(repoList, fmt.Sprintf("%s/%s: %s", *repo.Owner.Login, *repo.Name, *repo.Description))
    }

    return mcp.NewToolResultText(fmt.Sprintf("Found repositories:\n%s", repoList)), nil
}

func handleCreateIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    owner := request.Params.Arguments["owner"].(string)
    repo := request.Params.Arguments["repo"].(string)
    title := request.Params.Arguments["title"].(string)
    body := request.Params.Arguments["body"].(string)

    issue := &github.IssueRequest{
        Title: &title,
        Body:  &body,
    }

    result, _, err := githubClient.Issues.Create(ctx, owner, repo, issue)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Error creating issue: %v", err)), nil
    }

    return mcp.NewToolResultText(fmt.Sprintf("Created issue #%d: %s", *result.Number, *result.HTMLURL)), nil
}