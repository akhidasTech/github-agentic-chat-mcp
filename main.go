package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/akhidasTech/github-agentic-chat-mcp/pkg/vectorstore"
    "github.com/google/go-github/v57/github"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "golang.org/x/oauth2"
)

var (
    githubClient *github.Client
    vectorStore  *vectorstore.VectorStore
)

func init() {
    // Initialize GitHub client
    token := os.Getenv("GITHUB_TOKEN")
    if token == "" {
        log.Fatal("GITHUB_TOKEN environment variable is required")
    }

    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    githubClient = github.NewClient(tc)

    // Initialize vector store
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is required")
    }

    openaiKey := os.Getenv("OPENAI_API_KEY")
    if openaiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable is required")
    }

    var err error
    vectorStore, err = vectorstore.NewVectorStore(dbURL, openaiKey)
    if err != nil {
        log.Fatalf("Failed to initialize vector store: %v", err)
    }
}

func main() {
    // Create MCP server
    s := server.NewMCPServer(
        "GitHub Agentic Chat MCP",
        "1.0.0",
        server.WithResourceCapabilities(true, true),
        server.WithLogging(),
    )

    // Add vector search tools
    addVectorSearchTool := mcp.NewTool("add_to_vector_store",
        mcp.WithDescription("Add a document to the vector store"),
        mcp.WithString("content",
            mcp.Required(),
            mcp.Description("The content to store"),
        ),
        mcp.WithString("metadata",
            mcp.Required(),
            mcp.Description("JSON string of metadata"),
        ),
    )
    s.AddTool(addVectorSearchTool, handleAddToVectorStore)

    searchVectorTool := mcp.NewTool("vector_search",
        mcp.WithDescription("Search the vector store"),
        mcp.WithString("query",
            mcp.Required(),
            mcp.Description("The search query"),
        ),
        mcp.WithNumber("limit",
            mcp.Description("Maximum number of results"),
            mcp.Default(5),
        ),
    )
    s.AddTool(searchVectorTool, handleVectorSearch)

    // Add GitHub tools
    searchReposTool := mcp.NewTool("search_repositories",
        mcp.WithDescription("Search GitHub repositories"),
        mcp.WithString("query",
            mcp.Required(),
            mcp.Description("Search query for repositories"),
        ),
    )
    s.AddTool(searchReposTool, handleSearchRepositories)

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

func handleAddToVectorStore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    content := request.Params.Arguments["content"].(string)
    metadataStr := request.Params.Arguments["metadata"].(string)

    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Invalid metadata JSON: %v", err)), nil
    }

    if err := vectorStore.AddDocument(ctx, content, metadata); err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Failed to add document: %v", err)), nil
    }

    return mcp.NewToolResultText("Document added successfully"), nil
}

func handleVectorSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    query := request.Params.Arguments["query"].(string)
    limit := int(request.Params.Arguments["limit"].(float64))

    docs, err := vectorStore.Search(ctx, query, limit)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
    }

    var results []string
    for _, doc := range docs {
        results = append(results, fmt.Sprintf("Content: %s\nMetadata: %s\n---", doc.Content, doc.Metadata))
    }

    return mcp.NewToolResultText(fmt.Sprintf("Found %d documents:\n\n%s", len(docs), results)), nil
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

    return mcp.NewToolResultText(fmt.Sprintf("Found repositories:\n%v", repoList)), nil
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