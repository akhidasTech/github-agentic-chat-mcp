# GitHub Agentic Chat MCP Server

This is an MCP (Model Context Protocol) server implementation for GitHub agentic chat using Go. It provides tools for interacting with GitHub through natural language and includes vector search capabilities.

## Features

- Search GitHub repositories
- Create issues
- Vector search functionality
  - Add documents to vector store
  - Semantic search across stored documents
- Extensible structure for adding more features

## Prerequisites

- Go 1.21 or later
- PostgreSQL with pgvector extension
- GitHub Personal Access Token
- OpenAI API Key
- Claude Desktop or other MCP-compatible client

## Setup

1. Clone the repository:
```bash
git clone https://github.com/akhidasTech/github-agentic-chat-mcp.git
cd github-agentic-chat-mcp
```

2. Set up environment variables:
```bash
export GITHUB_TOKEN=your_github_token_here
export DATABASE_URL=postgres://user:password@localhost:5432/dbname
export OPENAI_API_KEY=your_openai_api_key_here
```

3. Set up PostgreSQL with pgvector:
```sql
CREATE EXTENSION vector;
```

4. Build the server:
```bash
go build -o bin/github-agentic-chat-mcp
```

5. Configure Claude Desktop:
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

6. Restart Claude Desktop

## Available Tools

### GitHub Tools

#### search_repositories
Search for GitHub repositories using a query string.

#### create_issue
Create a new issue in a GitHub repository.

### Vector Search Tools

#### add_to_vector_store
Add a document to the vector store for semantic search.

Parameters:
- content: The text content to store
- metadata: JSON string containing metadata about the content

Example:
```json
{
    "content": "This is a document about GitHub Actions",
    "metadata": "{\"type\": \"documentation\", \"tags\": [\"github\", \"ci-cd\"]}"
}
```

#### vector_search
Perform semantic search across stored documents.

Parameters:
- query: The search query text
- limit: Maximum number of results to return (default: 5)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License