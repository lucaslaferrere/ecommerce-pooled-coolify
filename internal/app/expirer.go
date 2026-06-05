package app

import (
	"context"
	"log"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/services"

	"go.mongodb.org/mongo-driver/mongo"
)

// StartOrderExpirer lanza una goroutine que cancela órdenes "pending" abandonadas
// y restaura su stock. Corre cada `interval` y vence órdenes más viejas que `maxAge`.
// Se detiene cuando `ctx` es cancelado.
func StartOrderExpirer(ctx context.Context, db *mongo.Database, interval, maxAge time.Duration) {
	orderRepo   := repositories.NewOrderRepositoryMongo(db.Collection("orders"))
	productRepo := repositories.NewProductRepositoryMongo(db.Collection("products"))
	svc         := services.NewOrderService(orderRepo, productRepo)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				n, err := svc.ExpireAbandonedOrders(runCtx, maxAge)
				cancel()
				if err != nil {
					log.Printf("expire orders: %v", err)
				} else if n > 0 {
					log.Printf("expire orders: %d orden(es) cancelada(s), stock restaurado", n)
				}
			}
		}
	}()
}
