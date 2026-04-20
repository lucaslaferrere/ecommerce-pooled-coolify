package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"
)

// ImageStorageService es una interfaz para almacenar imágenes
type ImageStorageService interface {
	// UploadImage carga una imagen y retorna la URL
	UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error)

	// DeleteImage elimina una imagen por su URL
	DeleteImage(ctx context.Context, imageURL string) error

	// GenerateImageURL genera una URL simulada para una imagen
	GenerateImageURL(filename string) string
}

// S3ImageStorage es una implementación stub de ImageStorageService usando S3/Object Storage
type S3ImageStorage struct {
	bucketName string
	region     string
	baseURL    string
}

// NewS3ImageStorage crea una nueva instancia de S3ImageStorage
func NewS3ImageStorage(bucketName string, region string, baseURL string) *S3ImageStorage {
	return &S3ImageStorage{
		bucketName: bucketName,
		region:     region,
		baseURL:    baseURL,
	}
}

// UploadImage carga una imagen y retorna la URL simulada
// En producción, esto interactuaría con AWS S3 o un servicio de almacenamiento similar
func (s *S3ImageStorage) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Validaciones
	if file == nil {
		return "", fmt.Errorf("archivo no puede ser nil")
	}

	// Validar tamaño (máximo 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return "", fmt.Errorf("archivo demasiado grande (máximo 10MB)")
	}

	// Validar extensión
	ext := filepath.Ext(file.Filename)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExtensions[ext] {
		return "", fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	// STUB: Simular la carga (en producción, aquí se subiría a S3)
	// Generar un nombre único para el archivo
	timestamp := time.Now().Unix()
	randomID := fmt.Sprintf("%d", timestamp)
	filename := fmt.Sprintf("products/%s%s", randomID, ext)

	// Generar URL simulada
	imageURL := s.GenerateImageURL(filename)

	return imageURL, nil
}

// DeleteImage elimina una imagen (stub)
// En producción, esto eliminaría el archivo de S3
func (s *S3ImageStorage) DeleteImage(ctx context.Context, imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("URL de imagen no puede ser vacía")
	}

	// STUB: Simular la eliminación
	// En producción, aquí se eliminaría el objeto de S3
	return nil
}

// GenerateImageURL genera una URL simulada para una imagen
func (s *S3ImageStorage) GenerateImageURL(filename string) string {
	// Simular una URL de S3 o CDN
	// Formato: https://bucket-name.s3.region.amazonaws.com/filename
	// O si se usa un CDN: https://cdn.example.com/filename

	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", s.baseURL, filename)
	}

	// URL por defecto (simulada)
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, filename)
}

// LocalImageStorage es una alternativa para desarrollo local que almacena imágenes en el filesystem
type LocalImageStorage struct {
	uploadDir string
	baseURL   string
}

// NewLocalImageStorage crea una nueva instancia de LocalImageStorage
func NewLocalImageStorage(uploadDir string, baseURL string) *LocalImageStorage {
	return &LocalImageStorage{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadImage carga una imagen localmente
func (l *LocalImageStorage) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", fmt.Errorf("archivo no puede ser nil")
	}

	// Validar tamaño (máximo 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return "", fmt.Errorf("archivo demasiado grande (máximo 10MB)")
	}

	// Validar extensión
	ext := filepath.Ext(file.Filename)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExtensions[ext] {
		return "", fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	// STUB: Simular la carga local
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("products/%d%s", timestamp, ext)

	// Generar URL
	imageURL := l.GenerateImageURL(filename)

	// TODO: En producción real, guardaría el archivo aquí
	// err := l.saveFile(uploadDir, filename, file)
	// if err != nil {
	//     return "", err
	// }

	return imageURL, nil
}

// DeleteImage elimina una imagen local
func (l *LocalImageStorage) DeleteImage(ctx context.Context, imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("URL de imagen no puede ser vacía")
	}

	// TODO: En producción real, eliminaría el archivo del filesystem
	// filePath := filepath.Join(l.uploadDir, imageFileName)
	// return os.Remove(filePath)

	return nil
}

// GenerateImageURL genera una URL simulada para una imagen local
func (l *LocalImageStorage) GenerateImageURL(filename string) string {
	if l.baseURL != "" {
		return fmt.Sprintf("%s/%s", l.baseURL, filename)
	}

	// URL por defecto para desarrollo local
	return fmt.Sprintf("/uploads/%s", filename)
}
