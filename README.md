# GitHub Agentic Chat MCP Server

This is an MCP (Model Context Protocol) server implementation for GitHub agentic chat using Go. It provides tools for interacting with GitHub through natural language.

## Features

- Search GitHub repositories
- Create issues
- More features coming soon!

## Prerequisites

- Go 1.21 or later
- GitHub Personal Access Token
- Claude Desktop or other MCP-compatible client

## Setup

1. Clone the repository:
```bash
git clone https://github.com/akhidasTech/github-agentic-chat-mcp.git
cd github-agentic-chat-mcp
```

2. Set up your GitHub token:
```bash
export GITHUB_TOKEN=your_github_token_here
```

3. Build the server:
```bash
go build -o bin/github-agentic-chat-mcp
```

4. Configure Claude Desktop:
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
    "mcpServers": {
        "github-chat": {
            "command": "/absolute/path/to/bin/github-agentic-chat-mcp"
        }
    }
}
```

5. Restart Claude Desktop

## Available Tools

### search_repositories
Search for GitHub repositories using a query string.

### create_issue
Create a new issue in a GitHub repository.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License