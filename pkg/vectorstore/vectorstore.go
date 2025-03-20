package vectorstore

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"

    "github.com/lib/pq"
    "github.com/pgvector/pgvector-go"
    "github.com/sashabaranov/go-openai"
)

type VectorStore struct {
    db        *sql.DB
    openaiAPI *openai.Client
}

type Document struct {
    ID        int64     `json:"id"`
    Content   string    `json:"content"`
    Metadata  string    `json:"metadata"`
    Embedding []float32 `json:"embedding"`
}

func NewVectorStore(dbURL string, openaiKey string) (*VectorStore, error) {
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }

    // Initialize pgvector extension
    if _, err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
        return nil, fmt.Errorf("failed to create vector extension: %v", err)
    }

    // Create documents table if not exists
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS documents (
            id SERIAL PRIMARY KEY,
            content TEXT NOT NULL,
            metadata JSONB,
            embedding vector(1536)
        )
    `)
    if err != nil {
        return nil, fmt.Errorf("failed to create documents table: %v", err)
    }

    return &VectorStore{
        db:        db,
        openaiAPI: openai.NewClient(openaiKey),
    }, nil
}

func (vs *VectorStore) AddDocument(ctx context.Context, content string, metadata map[string]interface{}) error {
    // Get embedding from OpenAI
    resp, err := vs.openaiAPI.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Input: []string{content},
        Model: openai.AdaEmbeddingV2,
    })
    if err != nil {
        return fmt.Errorf("failed to create embedding: %v", err)
    }

    if len(resp.Data) == 0 {
        return fmt.Errorf("no embedding returned from OpenAI")
    }

    // Convert metadata to JSON
    metadataJSON, err := json.Marshal(metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %v", err)
    }

    // Store document with embedding
    _, err = vs.db.ExecContext(ctx,
        "INSERT INTO documents (content, metadata, embedding) VALUES ($1, $2, $3)",
        content,
        metadataJSON,
        pgvector.NewVector(resp.Data[0].Embedding),
    )
    if err != nil {
        return fmt.Errorf("failed to insert document: %v", err)
    }

    return nil
}

func (vs *VectorStore) Search(ctx context.Context, query string, limit int) ([]Document, error) {
    // Get query embedding from OpenAI
    resp, err := vs.openaiAPI.CreateEmbeddings(ctx, openai.EmbeddingRequest{
        Input: []string{query},
        Model: openai.AdaEmbeddingV2,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create query embedding: %v", err)
    }

    if len(resp.Data) == 0 {
        return nil, fmt.Errorf("no embedding returned from OpenAI")
    }

    // Search for similar documents
    rows, err := vs.db.QueryContext(ctx,
        "SELECT id, content, metadata, embedding::float[] FROM documents ORDER BY embedding <-> $1 LIMIT $2",
        pgvector.NewVector(resp.Data[0].Embedding),
        limit,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to search documents: %v", err)
    }
    defer rows.Close()

    var documents []Document
    for rows.Next() {
        var doc Document
        var embeddingArray []float64
        if err := rows.Scan(&doc.ID, &doc.Content, &doc.Metadata, pq.Array(&embeddingArray)); err != nil {
            log.Printf("Error scanning row: %v", err)
            continue
        }
        // Convert float64 to float32
        doc.Embedding = make([]float32, len(embeddingArray))
        for i, v := range embeddingArray {
            doc.Embedding[i] = float32(v)
        }
        documents = append(documents, doc)
    }

    return documents, nil
}