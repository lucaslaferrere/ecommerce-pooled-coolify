package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client contiene la conexión a MongoDB
type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewClient crea una nueva conexión a MongoDB
func NewClient(ctx context.Context, mongoURI string, dbName string) (*Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Verificar la conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	return &Client{
		client: client,
		db:     db,
	}, nil
}

// GetDatabase retorna la instancia de la base de datos
func (c *Client) GetDatabase() *mongo.Database {
	return c.db
}

// GetClient retorna el cliente de MongoDB
func (c *Client) GetClient() *mongo.Client {
	return c.client
}

// Close cierra la conexión a MongoDB
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
