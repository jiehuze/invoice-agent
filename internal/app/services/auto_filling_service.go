package services

import (
	"fmt"
	"github.com/openai/openai-go"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

// 常量定义
const (
	DelayShort  = 1 * time.Second
	DelayNormal = 10 * time.Second
	DelayLong   = 15 * time.Second
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

type AutoFillingService struct {
	client openai.Client
}

var AutoFillingClient *AutoFillingService

func NewAutoFillingService() *AutoFillingService {
	if AutoFillingClient == nil {
		AutoFillingClient = &AutoFillingService{}
	}
	return AutoFillingClient
}

func AutoFillingServiceStart() {
	start()
}

func start() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("安装 Playwright 失败: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("启动 Playwright 失败: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("启动浏览器失败: %v", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("创建上下文失败: %v", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("创建页面失败: %v", err)
	}

	// 执行主要流程
	if err := getContent(page); err != nil {
		log.Fatalf("执行流程失败: %v", err)
	}
}

func getContent(page playwright.Page) error {
	// 导航到目标URL
	if _, err := page.Goto("http://open.sky-dome.com.cn:9086/"); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	// 执行登录操作
	if err := handleLogin(page); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	// 打开新增对话框
	if err := handleAddDialog(page); err != nil {
		return fmt.Errorf("打开新增对话框失败: %w", err)
	}

	// 填写基础信息
	basicItem := BasicItem{
		Category:   "日常报销",
		Title:      "主题xxxxxxxxxxxxxxxxx",
		UrgentType: "紧急",
		Comment:    "xxxxxxxxxxxxxxxxxxxxxxx项目出差",
	}
	if err := handleReimburseBasic(page, basicItem); err != nil {
		return fmt.Errorf("填写基础信息失败: %w", err)
	}

	if err := handleWheelDown(page, 0, 500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	// 填写支付信息
	payItem := PayItem{
		BusinessDept: "智能业务部/智能业务部-IT外包项目",
		BudgetDept:   "智能业务部/智能业务部-IT外包项目",
		PayDept:      "天宇正清科技有限公司",
		ProjectType:  "成本中心",
		Project:      "成本中心",
	}
	if err := handleReimbursePayInfo(page, payItem); err != nil {
		return fmt.Errorf("填写支付信息失败: %w", err)
	}

	// 填写报销明细
	item1 := CostItem{
		Category:   "团建费",
		Name:       "本部门团建",
		Comment:    "费用说明xxxxxxxxxxxxxxxxxx",
		Cost:       "1500",
		BillNumber: "3",
	}
	item2 := CostItem{
		Category:   "办公费",
		Name:       "办公用电",
		Comment:    "办公费用说明",
		Cost:       "800",
		BillNumber: "4",
	}

	for i := 0; i < 2; i++ {
		if err := handleAddDetail(page); err != nil {
			return fmt.Errorf("添加明细失败: %w", err)
		}
	}

	items := []CostItem{item1, item2}
	for i, item := range items {
		if err := handleReimburseDetail(page, item, i+1); err != nil {
			return fmt.Errorf("填写报销明细失败: %w", err)
		}
	}

	// 增值税发票批量上传
	if err := handleWheelDown(page, 0, 500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	if err := handleVatInvoiceUpload(page, "invoice2.pdf"); err != nil {
		return fmt.Errorf("上传发票失败: %w", err)
	}
	if err := handleVatInvoiceUpload(page, "invoice3.pdf"); err != nil {
		return fmt.Errorf("上传发票失败: %w", err)
	}

	if err := handleWheelDown(page, 0, -500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	if err := handleSaveInfo(page); err != nil {
		return fmt.Errorf("保存信息失败: %w", err)
	}

	return nil
}

func handleLogin(page playwright.Page) error {
	fmt.Println("进入登录页面...")

	// 等待用户名输入框
	if err := page.Locator("[placeholder=\"请输入账号\"]").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		return fmt.Errorf("等待用户名输入框失败: %w", err)
	}

	// 填写用户名
	if err := page.Locator("[placeholder=\"请输入账号\"]").Fill("tyzq-wangmeng5"); err != nil {
		return fmt.Errorf("填写用户名失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 填写密码
	if err := page.Locator("[placeholder=\"请输入密码\"]").Fill("tyzq123456"); err != nil {
		return fmt.Errorf("填写密码失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 点击登录按钮
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "登录"}).Click(); err != nil {
		return fmt.Errorf("点击登录按钮失败: %w", err)
	}

	// 等待并重新加载
	time.Sleep(DelayLong)

	if _, err := page.Goto("http://open.sky-dome.com.cn:9086/#/reimbursement/employee"); err != nil {
		return fmt.Errorf("导航到报销页面失败: %w", err)
	}

	time.Sleep(DelayNormal)
	fmt.Println("登录完成")
	return nil
}

func handleAddDialog(page playwright.Page) error {
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "新增"}).Click(); err != nil {
		return fmt.Errorf("点击新增按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("打开新增报销记录对话框")
	return nil
}

func handleReimburseBasic(page playwright.Page, basicItem BasicItem) error {
	// 设置报销类型
	if err := handleReimburseType(page, basicItem.Category); err != nil {
		return fmt.Errorf("设置报销类型失败: %w", err)
	}

	// 设置紧急类型
	if err := handleUrgentType(page, basicItem.UrgentType); err != nil {
		return fmt.Errorf("设置紧急类型失败: %w", err)
	}

	// 填写报销说明
	if err := handleReimburseComment(page, basicItem.Comment); err != nil {
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("基础信息填写完成")
	return nil
}

func handleReimburseType(page playwright.Page, category string) error {
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择报销类型\"]").Click(); err != nil {
		return fmt.Errorf("点击报销类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择报销类型
	if err := page.Locator(fmt.Sprintf("li.el-select-dropdown__item:has-text(\"%s\")", category)).Last().Click(); err != nil {
		return fmt.Errorf("选择报销类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("设置报销类型完成")
	return nil
}

func handleUrgentType(page playwright.Page, urgentType string) error {
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择紧急类型\"]").Click(); err != nil {
		return fmt.Errorf("点击紧急类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择紧急类型
	if err := page.GetByText(urgentType, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(); err != nil {
		return fmt.Errorf("选择紧急类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("设置紧急类型完成")
	return nil
}

func handleReimburseComment(page playwright.Page, comment string) error {
	if err := page.Locator("[placeholder=\"请输入报销说明\"]").Fill(comment); err != nil {
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("设置报销说明完成")
	return nil
}

func handleWheelDown(page playwright.Page, x float64, y float64) error {
	if err := page.Mouse().Wheel(x, y); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	return nil
}

func handleReimbursePayInfo(page playwright.Page, payItem PayItem) error {
	// 设置业务发生部门
	if err := handleBusinessDept(page, payItem.BusinessDept); err != nil {
		return fmt.Errorf("设置业务发生部门失败: %w", err)
	}

	// 设置预算承担部门
	if err := handleBudgetDept(page, payItem.BudgetDept); err != nil {
		return fmt.Errorf("设置预算承担部门失败: %w", err)
	}

	// 设置费用归集项目
	if err := handleProjectType(page, payItem.ProjectType); err != nil {
		return fmt.Errorf("设置项目类型失败: %w", err)
	}
	//if err := handleProject(page, payItem.Project); err != nil {
	//	return fmt.Errorf("设置项目失败: %w", err)
	//}

	// 设置付款公司
	if err := handlePayDept(page, payItem.PayDept); err != nil {
		return fmt.Errorf("设置付款部门失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("支付信息填写完成")
	return nil
}

func handleBusinessDept(page playwright.Page, businessDeptText string) error {
	businessDept := page.Locator("div.el-form-item.is-required.custom-form-render-item.custom-form-render-item-twoline")
	if err := businessDept.Locator("[placeholder='请选择']").Click(); err != nil {
		return fmt.Errorf("点击业务发生部门下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(businessDeptText)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择业务发生部门失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("设置业务发生部门完成")
	return nil
}

func handleBudgetDept(page playwright.Page, budgetDept string) error {
	budgetLocator := page.Locator("span:has-text(\"预算承担部门\")").Locator("..").Locator("..").Locator(".el-input__inner")
	if err := budgetLocator.Click(); err != nil {
		return fmt.Errorf("点击预算承担部门下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(budgetDept)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择预算承担部门失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("设置预算承担部门完成")
	return nil
}

func handleProjectType(page playwright.Page, projectType string) error {
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"项目类型\"]").Click(); err != nil {
		return fmt.Errorf("点击项目类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(projectType)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择项目类型失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("设置项目类型完成")
	return nil
}

func handleProject(page playwright.Page, project string) error {
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择项目\"]").Click(); err != nil {
		return fmt.Errorf("点击项目下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(project)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择项目失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("设置项目完成")
	return nil
}

func handlePayDept(page playwright.Page, payDept string) error {
	payLocator := page.Locator("span:has-text(\"付款公司\")").Locator("..").Locator("..").Locator(".el-input__inner")
	if err := payLocator.Click(); err != nil {
		return fmt.Errorf("点击付款公司下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(payDept)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择付款公司失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("设置付款部门完成")
	return nil
}

func handleAddDetail(page playwright.Page) error {
	addButton := page.Locator("button:has-text(\"导出\") + button")
	if err := addButton.Click(); err != nil {
		return fmt.Errorf("点击新增明细按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	fmt.Println("新增一条报销细节记录")
	return nil
}

func handleReimburseDetail(page playwright.Page, item CostItem, trIndex int) error {
	divElements := page.Locator("div.el-table__header-wrapper")
	count, err := divElements.Count()
	if err != nil {
		return fmt.Errorf("获取表格头数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		divElement := divElements.Nth(i)
		if textCount, _ := divElement.GetByText("费用名称").Count(); textCount > 0 {
			trElements := divElement.Locator("+ div").Locator("tr")
			trCount, err := trElements.Count()
			if err != nil {
				return fmt.Errorf("获取表格行数量失败: %w", err)
			}

			for j := 0; j < trCount; j++ {
				if j+1 == trIndex {
					tdElements := trElements.Nth(j).Locator("td")
					tdCount, err := tdElements.Count()
					if err != nil {
						return fmt.Errorf("获取表格列数量失败: %w", err)
					}

					for k := 0; k < tdCount; k++ {
						tdElement := tdElements.Nth(k)
						switch k {
						case 1: // 费用类别
							if err := handleCostCategoryInDetail(page, tdElement, item.Category); err != nil {
								return fmt.Errorf("设置费用类别失败: %w", err)
							}
						case 2: // 费用名称
							if err := handleCostNameInDetail(page, tdElement, item.Name); err != nil {
								return fmt.Errorf("设置费用名称失败: %w", err)
							}
						case 3: // 费用说明
							if err := handleCostCommentInDetail(page, tdElement, item.Comment); err != nil {
								return fmt.Errorf("设置费用说明失败: %w", err)
							}
						case 4: // 报销金额
							if err := handleCostInDetail(page, tdElement, item.Cost); err != nil {
								return fmt.Errorf("设置报销金额失败: %w", err)
							}
						case 5: // 发票张数
							if err := handleBillNumberInDetail(page, tdElement, item.BillNumber); err != nil {
								return fmt.Errorf("设置发票张数失败: %w", err)
							}
						}
					}
					break
				}
			}
			break
		}
	}

	time.Sleep(DelayShort)
	fmt.Println("报销明细填写完成")
	return nil
}

func handleCostCategoryInDetail(page playwright.Page, tdElement playwright.Locator, costCategory string) error {
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		return fmt.Errorf("点击费用类别输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allCategoryItems := page.GetByText(costCategory)
	count, err := allCategoryItems.Count()
	if err != nil {
		return fmt.Errorf("获取费用类别选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		categoryItem := allCategoryItems.Nth(i)
		visible, err := categoryItem.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			if err := categoryItem.Click(); err != nil {
				return fmt.Errorf("选择费用类别失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	return nil
}

func handleCostNameInDetail(page playwright.Page, tdElement playwright.Locator, costName string) error {
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		return fmt.Errorf("点击费用名称输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allNameItems := page.GetByText(costName)
	count, err := allNameItems.Count()
	if err != nil {
		return fmt.Errorf("获取费用名称选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		nameItem := allNameItems.Nth(i)
		visible, err := nameItem.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			if err := nameItem.Click(); err != nil {
				return fmt.Errorf("选择费用名称失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	return nil
}

func handleCostCommentInDetail(page playwright.Page, tdElement playwright.Locator, costComment string) error {
	if err := tdElement.Locator("input.el-input__inner").Fill(costComment); err != nil {
		return fmt.Errorf("填写费用说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	return nil
}

func handleCostInDetail(page playwright.Page, tdElement playwright.Locator, cost string) error {
	if err := tdElement.Locator("input.el-input__inner").Fill(cost); err != nil {
		return fmt.Errorf("填写报销金额失败: %w", err)
	}

	time.Sleep(DelayShort)
	return nil
}

func handleBillNumberInDetail(page playwright.Page, tdElement playwright.Locator, billNumber string) error {
	if err := tdElement.Locator("input.el-input__inner").Fill(billNumber); err != nil {
		return fmt.Errorf("填写发票张数失败: %w", err)
	}

	time.Sleep(DelayShort)
	return nil
}

func handleVatInvoiceUpload(page playwright.Page, fileName string) error {
	if err := page.Locator("input.el-upload__input").SetInputFiles([]string{fileName}); err != nil {
		return fmt.Errorf("上传发票文件失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	return nil
}

func handleSaveInfo(page playwright.Page) error {
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "保存"}).Click(); err != nil {
		return fmt.Errorf("点击保存按钮失败: %w", err)
	}

	time.Sleep(DelayNormal)
	return nil
}
