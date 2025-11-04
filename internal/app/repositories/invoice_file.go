// internal/app/repositories/invoice_file_repository.go
package repositories

import (
	"fmt"
	"invoice-agent/internal/app/models"
	"invoice-agent/internal/pkg/storage"
)

type InvoiceFileRepository struct{}

func NewInvoiceFileRepository() *InvoiceFileRepository {
	return &InvoiceFileRepository{}
}

// Create 创建新的发票文件记录
func (r *InvoiceFileRepository) Create(invoiceFile *models.InvoiceFile) error {
	//if err := invoiceFile.Validate(); err != nil {
	//	return err
	//}
	return storage.DB.Create(invoiceFile).Error
}

// CreateBatch 批量创建发票文件记录
func (r *InvoiceFileRepository) CreateBatch(invoiceFiles []models.InvoiceFile) error {
	fmt.Print("2----------------------\n")
	// 验证所有记录
	for i := range invoiceFiles {
		if err := invoiceFiles[i].Validate(); err != nil {
			return err
		}
	}

	fmt.Print("1----------------------\n")

	return storage.DB.CreateInBatches(invoiceFiles, 100).Error
}

// Update 更新发票文件记录（支持部分字段更新）
func (r *InvoiceFileRepository) Update(id uint64, invoiceFile *models.InvoiceFile) error {
	// 创建一个map来存储非零值字段
	updates := make(map[string]interface{})

	// 只更新有值的字段，避免覆盖原有数据
	if invoiceFile.InvoiceType != "" {
		updates["invoice_type"] = invoiceFile.InvoiceType
	}
	if invoiceFile.InvoiceCode != "" {
		updates["invoice_code"] = invoiceFile.InvoiceCode
	}
	if invoiceFile.FileName != "" {
		updates["file_name"] = invoiceFile.FileName
	}
	if invoiceFile.FileID != "" {
		updates["file_id"] = invoiceFile.FileID
	}
	if invoiceFile.IssueDate != "" {
		updates["issue_date"] = invoiceFile.IssueDate
	}
	if invoiceFile.BuyerName != "" {
		updates["buyer_name"] = invoiceFile.BuyerName
	}
	if invoiceFile.SellerName != "" {
		updates["seller_name"] = invoiceFile.SellerName
	}
	if invoiceFile.ItemName != "" {
		updates["item_name"] = invoiceFile.ItemName
	}
	if invoiceFile.ExpenseCategory != "" {
		updates["expense_category"] = invoiceFile.ExpenseCategory
	}

	// 数值类型字段更新
	if invoiceFile.Amount != 0 {
		updates["amount"] = invoiceFile.Amount
	}
	if invoiceFile.TaxAmount != 0 {
		updates["tax_amount"] = invoiceFile.TaxAmount
	}
	if invoiceFile.TotalAmount != 0 {
		updates["total_amount"] = invoiceFile.TotalAmount
	}

	// 枚举类型字段更新
	if invoiceFile.ServiceType != 0 {
		updates["service_type"] = invoiceFile.ServiceType
	}

	// 执行更新
	return storage.DB.Model(&models.InvoiceFile{}).Where("id = ?", id).Updates(updates).Error
}

func (r *InvoiceFileRepository) UpdateByFileId(fileId string, invoiceFile *models.InvoiceFile) error {
	// 创建一个map来存储非零值字段
	updates := make(map[string]interface{})

	// 只更新有值的字段，避免覆盖原有数据
	if invoiceFile.InvoiceType != "" {
		updates["invoice_type"] = invoiceFile.InvoiceType
	}
	if invoiceFile.InvoiceCode != "" {
		updates["invoice_code"] = invoiceFile.InvoiceCode
	}
	if invoiceFile.FileName != "" {
		updates["file_name"] = invoiceFile.FileName
	}
	if invoiceFile.FileID != "" {
		updates["file_id"] = invoiceFile.FileID
	}
	if invoiceFile.IssueDate != "" {
		updates["issue_date"] = invoiceFile.IssueDate
	}
	if invoiceFile.BuyerName != "" {
		updates["buyer_name"] = invoiceFile.BuyerName
	}
	if invoiceFile.SellerName != "" {
		updates["seller_name"] = invoiceFile.SellerName
	}
	if invoiceFile.ItemName != "" {
		updates["item_name"] = invoiceFile.ItemName
	}
	if invoiceFile.ExpenseCategory != "" {
		updates["expense_category"] = invoiceFile.ExpenseCategory
	}

	// 数值类型字段更新
	if invoiceFile.Amount != 0 {
		updates["amount"] = invoiceFile.Amount
	}
	if invoiceFile.TaxAmount != 0 {
		updates["tax_amount"] = invoiceFile.TaxAmount
	}
	if invoiceFile.TotalAmount != 0 {
		updates["total_amount"] = invoiceFile.TotalAmount
	}

	// 枚举类型字段更新
	if invoiceFile.ServiceType != 0 {
		updates["service_type"] = invoiceFile.ServiceType
	}

	// 执行更新
	return storage.DB.Model(&models.InvoiceFile{}).Where("file_id = ?", fileId).Updates(updates).Error
}

// GetByID 根据ID获取发票文件记录
func (r *InvoiceFileRepository) GetByID(id uint64) (*models.InvoiceFile, error) {
	var invoiceFile models.InvoiceFile
	err := storage.DB.Where("id = ?", id).First(&invoiceFile).Error
	if err != nil {
		return nil, err
	}
	return &invoiceFile, nil
}

// GetByFileID 根据文件ID获取发票文件记录
func (r *InvoiceFileRepository) GetByFileID(fileID string) (*models.InvoiceFile, error) {
	var invoiceFile models.InvoiceFile
	err := storage.DB.Where("file_id = ?", fileID).First(&invoiceFile).Error
	if err != nil {
		return nil, err
	}
	return &invoiceFile, nil
}

// List 获取发票文件列表
func (r *InvoiceFileRepository) List(limit, offset int) ([]models.InvoiceFile, error) {
	var invoiceFiles []models.InvoiceFile
	err := storage.DB.Limit(limit).Offset(offset).Find(&invoiceFiles).Error
	return invoiceFiles, err
}

func (r *InvoiceFileRepository) ListByCont(invoice models.InvoiceFile, limit, offset int) ([]models.InvoiceFile, error) {
	var invoiceFiles []models.InvoiceFile
	db := storage.DB

	if invoice.InvoiceType != "" {
		db = db.Where("invoice_type = ?", invoice.InvoiceType)
	}
	if invoice.InvoiceCode != "" {
		db = db.Where("invoice_code = ?", invoice.InvoiceCode)
	}
	if invoice.DocumentNumber != "" {
		db = db.Where("document_number = ?", invoice.DocumentNumber)
	}
	if invoice.FileName != "" {
		db = db.Where("file_name = ?", invoice.FileName)
	}
	if invoice.FileID != "" {
		db = db.Where("file_id = ?", invoice.FileID)
	}
	if invoice.IssueDate != "" {
		db = db.Where("issue_date = ?", invoice.IssueDate)
	}
	if invoice.BuyerName != "" {
		db = db.Where("buyer_name = ?", invoice.BuyerName)
	}
	if invoice.SellerName != "" {
		db = db.Where("seller_name = ?", invoice.SellerName)
	}
	if invoice.ItemName != "" {
		db = db.Where("item_name = ?", invoice.ItemName)
	}
	if invoice.ExpenseCategory != "" {
		db = db.Where("expense_category = ?", invoice.ExpenseCategory)
	}

	err := db.Limit(limit).Offset(offset).Find(&invoiceFiles).Error
	return invoiceFiles, err
}

// Delete 软删除发票文件记录
func (r *InvoiceFileRepository) Delete(id uint64) error {
	return storage.DB.Where("id = ?", id).Delete(&models.InvoiceFile{}).Error
}
