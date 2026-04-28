package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client contiene la conexión a MongoDB
type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewClient crea una nueva conexión a MongoDB.
// Las credenciales se extraen del URI y se pasan explícitamente al driver
// mediante options.Credential para evitar problemas de parseo del string.
func NewClient(ctx context.Context, mongoURI string, dbName string) (*Client, error) {
	creds, err := credentialsFromURI(mongoURI)
	if err != nil {
		return nil, fmt.Errorf("parseando credenciales de MONGO_URI: %w", err)
	}

	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetAuth(creds)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// credentialsFromURI extrae usuario, contraseña y authSource del URI de MongoDB.
func credentialsFromURI(rawURI string) (options.Credential, error) {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return options.Credential{}, err
	}

	cred := options.Credential{
		AuthSource: parsed.Query().Get("authSource"),
		Username:   parsed.User.Username(),
	}
	if p, ok := parsed.User.Password(); ok {
		cred.Password = p
	}
	// authSource por defecto es "admin" para conexiones con credenciales
	if cred.AuthSource == "" && cred.Username != "" {
		cred.AuthSource = "admin"
	}
	return cred, nil
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
