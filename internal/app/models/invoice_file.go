package models

import (
	"fmt"
	"time"
)

// ServiceType 服务类型枚举
type ServiceType int8

const (
	ServiceTypeInvoice    ServiceType = 1 // 发票类
	ServiceTypeNonInvoice ServiceType = 2 // 非发票类
)

// ExpenseCategory 费用类别枚举
type ExpenseCategory string

const (
	ExpenseCategoryTransport     ExpenseCategory = "市内交通费"
	ExpenseCategoryEntertainment ExpenseCategory = "业务招待费"
	ExpenseCategoryTravel        ExpenseCategory = "差旅费"
)

// 数据结构定义
type CostItem struct {
	Category   string
	Name       string
	Comment    string
	Cost       string
	BillNumber string
}

type BasicItem struct {
	Category   string
	Title      string
	UrgentType string
	Comment    string
}

type PayItem struct {
	BusinessDept string
	BudgetDept   string
	PayDept      string
	ProjectType  string
	Project      string
}

type AutoFillingRequest struct {
	BasicInfo BasicItem  `json:"basic_info"`
	PayInfo   PayItem    `json:"pay_info"`
	CostItems []CostItem `json:"cost_items"`
	FilePaths []string   `json:"file_paths"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
}

// ExpenseCategoryStats 费用类别统计结果
type ExpenseCategoryStats struct {
	CategoryName string        `json:"category_name"` // 费用类别名称
	TotalAmount  float64       `json:"total_amount"`  // 该类别下发票类型的总金额
	TicketCount  int           `json:"ticket_count"`  // 该类别下票据总数
	Invoices     []InvoiceFile `json:"invoices"`      // 该类别下的所有票据
}

// StatByExpenseCategory 按费用类别统计发票文件
func StatByExpenseCategory(invoices []InvoiceFile) map[string]*ExpenseCategoryStats {
	stats := make(map[string]*ExpenseCategoryStats)

	// 遍历所有发票文件
	for _, invoice := range invoices {
		category := string(invoice.ExpenseCategory)

		// 如果该类别还没有统计信息，创建新的
		if _, exists := stats[category]; !exists {
			stats[category] = &ExpenseCategoryStats{
				CategoryName: category,
				TotalAmount:  0.0,
				TicketCount:  0,
				Invoices:     make([]InvoiceFile, 0),
			}
		}

		// 累计票据数
		stats[category].TicketCount++

		// 如果是发票类型（service_type为1），累计金额
		if invoice.ServiceType == ServiceTypeInvoice {
			stats[category].TotalAmount += invoice.TotalAmount
		}

		// 添加票据到该类别下
		stats[category].Invoices = append(stats[category].Invoices, invoice)
	}

	return stats
}

type InvoiceFile struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement;comment:主键" json:"id"`
	SessionId       string          `gorm:"size:100;default:null;comment:报销单据编号" json:"session_id"`
	InvoiceType     string          `gorm:"size:100;default:null;comment:票据类型" json:"invoice_type"`
	InvoiceCode     string          `gorm:"size:100;default:null;comment:发票号码" json:"invoice_code"`
	IssueDate       string          `gorm:"size:20;default:null;comment:开票日期" json:"issue_date"`
	ServiceType     ServiceType     `gorm:"not null;comment:服务类型：1=发票类，2=非发票类" json:"service_type"`
	Amount          float64         `gorm:"type:decimal(10,2);default:null;comment:金额（不含税）" json:"amount"`
	TaxAmount       float64         `gorm:"type:decimal(10,2);default:null;comment:税额" json:"tax_amount"`
	TotalAmount     float64         `gorm:"type:decimal(10,2);not null;comment:合计金额" json:"total_amount"`
	BuyerName       string          `gorm:"size:100;default:null;comment:购买方名称" json:"buyer_name"`
	BuyerID         string          `gorm:"size:100;default:null;comment:购买方识别号" json:"buyer_id"`
	SellerName      string          `gorm:"size:100;default:null;comment:销售方名称" json:"seller_name"`
	SellerID        string          `gorm:"size:100;default:null;comment:销售方识别号" json:"seller_id"`
	ItemName        string          `gorm:"size:100;default:null;comment:项目名称" json:"item_name"`
	ExpenseCategory ExpenseCategory `gorm:"size:100;default:null;comment:费用类别" json:"expense_category"`
	FileName        string          `gorm:"size:500;not null;comment:原始文件名" json:"file_name"`
	FilePath        string          `gorm:"size:1000;default:null;comment:文件路径" json:"file_path"` // 记录文件路径
	FileID          string          `gorm:"size:100;not null;uniqueIndex;comment:文件唯一ID" json:"file_id"`
	CreatedAt       time.Time       `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:记录创建时间" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:最后更新时间" json:"updated_at"`
}

// TableName 指定表名
func (InvoiceFile) TableName() string {
	return "invoice_file"
}

// ServiceTypeName 获取服务类型名称
func (s ServiceType) Name() string {
	switch s {
	case ServiceTypeInvoice:
		return "发票类"
	case ServiceTypeNonInvoice:
		return "非发票类"
	default:
		return "未知"
	}
}

// Validate 验证模型数据
func (i *InvoiceFile) Validate() error {
	if i.FileName == "" {
		return fmt.Errorf("文件名不能为空")
	}
	if i.FileID == "" {
		return fmt.Errorf("文件ID不能为空")
	}
	if i.TotalAmount <= 0 {
		return fmt.Errorf("合计金额必须大于0")
	}
	return nil
}
