// internal/app/services/invoice_file.go
package services

import (
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/app/repositories"
)

type IInvoiceFile interface {
	CreateInvoiceFile(invoiceFile *models.InvoiceFile) error
	CreateInvoiceFilesBatch(invoiceFiles []models.InvoiceFile) error
	UpdateInvoiceFile(id uint64, invoiceFile *models.InvoiceFile) error
	GetInvoiceFileByID(id uint64) (*models.InvoiceFile, error)
	GetInvoiceFileByFileID(fileID string) (*models.InvoiceFile, error)
	ListInvoiceFiles(limit, offset int) ([]models.InvoiceFile, error)
	ListInvoiceFilesByCont(invoiceFile models.InvoiceFile, limit, offset int) ([]models.InvoiceFile, error)
	DeleteInvoiceFile(id uint64) error
}
type InvoiceFileService struct {
	repo *repositories.InvoiceFileRepository
}

func NewInvoiceFileService() IInvoiceFile {
	return &InvoiceFileService{
		repo: repositories.NewInvoiceFileRepository(),
	}
}

func (s *InvoiceFileService) CreateInvoiceFile(invoiceFile *models.InvoiceFile) error {
	return s.repo.Create(invoiceFile)
}

func (s *InvoiceFileService) CreateInvoiceFilesBatch(invoiceFiles []models.InvoiceFile) error {
	return s.repo.CreateBatch(invoiceFiles)
}

func (s *InvoiceFileService) UpdateInvoiceFile(id uint64, invoiceFile *models.InvoiceFile) error {
	return s.repo.Update(id, invoiceFile)
}

func (s *InvoiceFileService) GetInvoiceFileByID(id uint64) (*models.InvoiceFile, error) {
	return s.repo.GetByID(id)
}

func (s *InvoiceFileService) GetInvoiceFileByFileID(fileID string) (*models.InvoiceFile, error) {
	return s.repo.GetByFileID(fileID)
}

func (s *InvoiceFileService) ListInvoiceFiles(limit, offset int) ([]models.InvoiceFile, error) {
	return s.repo.List(limit, offset)
}

func (s *InvoiceFileService) ListInvoiceFilesByCont(invoiceFile models.InvoiceFile, limit, offset int) ([]models.InvoiceFile, error) {
	return s.repo.ListByCont(invoiceFile, limit, offset)
}

func (s *InvoiceFileService) DeleteInvoiceFile(id uint64) error {
	return s.repo.Delete(id)
}
