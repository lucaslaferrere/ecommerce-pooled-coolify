package services

import (
	"context"
	"errors"
	"log"
	"time"

	"ecommerce-pooled/internal/adapters/repositories"
	"ecommerce-pooled/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvoiceNotFound = errors.New("factura no encontrada")

type InvoiceService struct {
	repo         *repositories.InvoiceRepositoryMongo
	orderService *OrderService
	emailService *EmailService
}

func NewInvoiceService(repo *repositories.InvoiceRepositoryMongo, orderService *OrderService, emailService *EmailService) *InvoiceService {
	return &InvoiceService{repo: repo, orderService: orderService, emailService: emailService}
}

// SaveAndSend guarda (o reemplaza) el PDF de la factura y lo envía al cliente.
// Primero persiste el PDF; si el envío falla, la factura queda guardada para
// reintentar con Resend. Solo marca sent_at cuando el envío fue exitoso.
func (s *InvoiceService) SaveAndSend(ctx context.Context, orderID primitive.ObjectID, filename, pdfBase64 string) (time.Time, error) {
	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if order == nil {
		return time.Time{}, ErrOrderNotFound
	}

	// Guardar/reemplazar el PDF (sin sent_at todavía).
	inv := &domain.Invoice{OrderID: orderID, Filename: filename, Content: pdfBase64, SentAt: nil}
	if err := s.repo.Upsert(ctx, inv); err != nil {
		return time.Time{}, err
	}

	// Enviar. Si falla, la factura queda guardada.
	if err := s.emailService.SendInvoice(order, pdfBase64, filename); err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	if err := s.repo.SetSentAt(ctx, orderID, now); err != nil {
		log.Printf("invoice: no se pudo marcar sent_at en la factura del pedido %s: %v", orderID.Hex(), err)
	}
	if err := s.orderService.SetInvoiceSentAt(ctx, orderID, now); err != nil {
		log.Printf("invoice: no se pudo marcar invoice_sent_at en el pedido %s: %v", orderID.Hex(), err)
	}
	return now, nil
}

// Resend reenvía la factura ya guardada del pedido.
func (s *InvoiceService) Resend(ctx context.Context, orderID primitive.ObjectID) (time.Time, error) {
	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if order == nil {
		return time.Time{}, ErrOrderNotFound
	}
	inv, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return time.Time{}, err
	}
	if inv == nil {
		return time.Time{}, ErrInvoiceNotFound
	}
	if err := s.emailService.SendInvoice(order, inv.Content, inv.Filename); err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	if err := s.repo.SetSentAt(ctx, orderID, now); err != nil {
		log.Printf("invoice: no se pudo marcar sent_at en la factura del pedido %s: %v", orderID.Hex(), err)
	}
	if err := s.orderService.SetInvoiceSentAt(ctx, orderID, now); err != nil {
		log.Printf("invoice: no se pudo marcar invoice_sent_at en el pedido %s: %v", orderID.Hex(), err)
	}
	return now, nil
}
