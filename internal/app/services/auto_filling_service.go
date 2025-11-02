package services

import (
	"fmt"
	"github.com/openai/openai-go"
	log "github.com/sirupsen/logrus"
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
	client   openai.Client
	progress chan string
}

var AutoFillingClient *AutoFillingService

// 封装发送进度信息的方法
func sendProgress(progress chan string, message string) {
	if progress != nil {
		select {
		case progress <- message + "\n":
		default:
		}
	}
	log.Info(message) //保留原有的控制台输出
}

func NewAutoFillingService() *AutoFillingService {
	if AutoFillingClient == nil {
		AutoFillingClient = &AutoFillingService{}
	}
	return AutoFillingClient
}

func AutoFillingServiceStart(progress chan string) {
	startWithProgress(progress)
}

func startWithProgress(progress chan string) {
	// 发送初始化信息
	sendProgress(progress, "开始初始化 Playwright...")

	err := playwright.Install()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("安装 Playwright 失败: %v", err))
		return
	}
	sendProgress(progress, "Playwright 安装完成")

	pw, err := playwright.Run()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("启动 Playwright 失败: %v", err))
		return
	}
	defer pw.Stop()
	sendProgress(progress, "Playwright 启动成功")

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		sendProgress(progress, fmt.Sprintf("启动浏览器失败: %v", err))
		return
	}
	defer browser.Close()
	sendProgress(progress, "浏览器启动成功")

	context, err := browser.NewContext()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("创建上下文失败: %v", err))
		return
	}
	defer context.Close()
	sendProgress(progress, "浏览器上下文创建成功")

	page, err := context.NewPage()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("创建页面失败: %v", err))
		return
	}
	sendProgress(progress, "新页面创建成功")

	// 执行主要流程
	if err := getContentWithProgress(page, progress); err != nil {
		sendProgress(progress, fmt.Sprintf("执行流程失败: %v", err))
		return
	}

	sendProgress(progress, "自动化填报流程已完成")
	close(progress) // 关闭通道表示结束
}

func getContentWithProgress(page playwright.Page, progress chan string) error {
	// 导航到目标URL
	sendProgress(progress, "正在导航到目标网址...")
	if _, err := page.Goto("http://open.sky-dome.com.cn:9086/"); err != nil {
		sendProgress(progress, fmt.Sprintf("导航失败: %v", err))
		return fmt.Errorf("导航失败: %w", err)
	}

	// 执行登录操作
	sendProgress(progress, "开始执行登录操作...")
	if err := handleLogin(page, progress); err != nil {
		sendProgress(progress, fmt.Sprintf("登录失败: %v", err))
		return fmt.Errorf("登录失败: %w", err)
	}

	// 打开新增对话框
	sendProgress(progress, "正在打开新增对话框...")
	if err := handleAddDialog(page, progress); err != nil {
		sendProgress(progress, fmt.Sprintf("打开新增对话框失败:  %v", err))
		return fmt.Errorf("打开新增对话框失败: %w", err)
	}

	// 填写基础信息
	basicItem := BasicItem{
		Category:   "日常报销",
		Title:      "主题xxxxxxxxxxxxxxxxx",
		UrgentType: "紧急",
		Comment:    "xxxxxxxxxxxxxxxxxxxxxxx项目出差",
	}
	if err := handleReimburseBasic(page, progress, basicItem); err != nil {
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
	if err := handleReimbursePayInfo(page, progress, payItem); err != nil {
		sendProgress(progress, fmt.Sprintf("填写支付信息失败:  %v", err))
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
		if err := handleAddDetail(page, progress); err != nil {
			return fmt.Errorf("添加明细失败: %w", err)
		}
	}

	items := []CostItem{item1, item2}
	for i, item := range items {
		if err := handleReimburseDetail(page, progress, item, i+1); err != nil {
			sendProgress(progress, fmt.Sprintf("填写报销明细失败:  %v", err))
			return fmt.Errorf("填写报销明细失败: %w", err)
		}
	}

	// 增值税发票批量上传
	if err := handleWheelDown(page, 0, 500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	if err := handleVatInvoiceUpload(page, progress, "invoice2.pdf"); err != nil {
		sendProgress(progress, fmt.Sprintf("上传发票失败:  %v", err))
		return fmt.Errorf("上传发票失败: %w", err)
	}
	if err := handleVatInvoiceUpload(page, progress, "invoice3.pdf"); err != nil {
		sendProgress(progress, fmt.Sprintf("上传发票失败:  %v", err))
		return fmt.Errorf("上传发票失败: %w", err)
	}

	if err := handleWheelDown(page, 0, -500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	if err := handleSaveInfo(page, progress); err != nil {
		sendProgress(progress, fmt.Sprintf("保存信息失败:  %v", err))
		return fmt.Errorf("保存信息失败: %w", err)
	}

	return nil
}

func handleLogin(page playwright.Page, progress chan string) error {
	sendProgress(progress, "进入登录页面...")

	// 等待用户名输入框
	sendProgress(progress, "等待输入用户名和密码...")
	if err := page.Locator("[placeholder=\"请输入账号\"]").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		sendProgress(progress, fmt.Sprintf("等待用户名输入框失败:  %v", err))
		return fmt.Errorf("等待用户名输入框失败: %w", err)
	}

	// 填写用户名
	if err := page.Locator("[placeholder=\"请输入账号\"]").Fill("tyzq-wangmeng5"); err != nil {
		sendProgress(progress, fmt.Sprintf("填写用户名失败:  %v", err))
		return fmt.Errorf("填写用户名失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 填写密码
	if err := page.Locator("[placeholder=\"请输入密码\"]").Fill("tyzq123456"); err != nil {
		sendProgress(progress, fmt.Sprintf("填写密码失败:  %v", err))
		return fmt.Errorf("填写密码失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 点击登录按钮
	sendProgress(progress, "登录系统...")
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "登录"}).Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击登录按钮失败:  %v", err))
		return fmt.Errorf("点击登录按钮失败: %w", err)
	}

	// 等待并重新加载
	time.Sleep(DelayLong)
	sendProgress(progress, "导航到报销页面...")
	if _, err := page.Goto("http://open.sky-dome.com.cn:9086/#/reimbursement/employee"); err != nil {
		sendProgress(progress, fmt.Sprintf("导航到报销页面失败:  %v", err))
		return fmt.Errorf("导航到报销页面失败: %w", err)
	}

	time.Sleep(DelayNormal)
	sendProgress(progress, "进入到报销页面...")
	return nil
}

func handleAddDialog(page playwright.Page, progress chan string) error {
	sendProgress(progress, "打开新增报销记录对话框...")
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "新增"}).Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击新增按钮失败:  %v", err))
		return fmt.Errorf("点击新增按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "打开新增报销记录对话框完成")
	return nil
}

func handleReimburseBasic(page playwright.Page, progress chan string, basicItem BasicItem) error {
	// 设置报销类型
	sendProgress(progress, "设置报销类型...")
	if err := handleReimburseType(page, progress, basicItem.Category); err != nil {
		sendProgress(progress, fmt.Sprintf("设置报销类型失败:  %v", err))
		return fmt.Errorf("设置报销类型失败: %w", err)
	}

	// 设置紧急类型
	sendProgress(progress, "设置紧急类型...")
	if err := handleUrgentType(page, progress, basicItem.UrgentType); err != nil {
		sendProgress(progress, fmt.Sprintf("设置紧急类型失败: %v", err))
		return fmt.Errorf("设置紧急类型失败: %w", err)
	}

	// 填写报销说明
	sendProgress(progress, "填写报销说明...")
	if err := handleReimburseComment(page, progress, basicItem.Comment); err != nil {
		sendProgress(progress, fmt.Sprintf("填写报销说明失败: %v", err))
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "基础信息填写完成...")
	return nil
}

func handleReimburseType(page playwright.Page, progress chan string, category string) error {
	sendProgress(progress, "填写报销类型...")
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择报销类型\"]").Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击报销类型下拉框失败: %v", err))
		return fmt.Errorf("点击报销类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择报销类型
	if err := page.Locator(fmt.Sprintf("li.el-select-dropdown__item:has-text(\"%s\")", category)).Last().Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("选择报销类型失败: %v", err))
		return fmt.Errorf("选择报销类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "设置报销类型完成...")
	return nil
}

func handleUrgentType(page playwright.Page, progress chan string, urgentType string) error {
	sendProgress(progress, "设置紧急类型...")
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择紧急类型\"]").Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击紧急类型下拉框失败: %v", err))
		return fmt.Errorf("点击紧急类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择紧急类型
	if err := page.GetByText(urgentType, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("选择紧急类型失败: %v", err))
		return fmt.Errorf("选择紧急类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "设置紧急类型完成...")
	return nil
}

func handleReimburseComment(page playwright.Page, progress chan string, comment string) error {
	sendProgress(progress, "填写报销说明:"+comment)
	if err := page.Locator("[placeholder=\"请输入报销说明\"]").Fill(comment); err != nil {
		sendProgress(progress, fmt.Sprintf("填写报销说明失败: %v", err))
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "设置报销说明完成...")
	return nil
}

func handleWheelDown(page playwright.Page, x float64, y float64) error {
	if err := page.Mouse().Wheel(x, y); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	return nil
}

func handleReimbursePayInfo(page playwright.Page, progress chan string, payItem PayItem) error {
	// 设置业务发生部门
	sendProgress(progress, "设置业务发生部门: "+payItem.BusinessDept)
	if err := handleBusinessDept(page, progress, payItem.BusinessDept); err != nil {
		sendProgress(progress, fmt.Sprintf("设置业务发生部门失败: %v", err))
		return fmt.Errorf("设置业务发生部门失败: %w", err)
	}

	// 设置预算承担部门
	sendProgress(progress, "设置业务发生部门: "+payItem.BudgetDept)
	if err := handleBudgetDept(page, progress, payItem.BudgetDept); err != nil {
		sendProgress(progress, fmt.Sprintf("设置预算承担部门失败: %v", err))
		return fmt.Errorf("设置预算承担部门失败: %w", err)
	}

	// 设置费用归集项目
	sendProgress(progress, "设置业务发生部门: "+payItem.ProjectType)
	if err := handleProjectType(page, progress, payItem.ProjectType); err != nil {
		sendProgress(progress, fmt.Sprintf("设置项目类型失败: %v", err))
		return fmt.Errorf("设置项目类型失败: %w", err)
	}
	//if err := handleProject(page, payItem.Project); err != nil {
	//	return fmt.Errorf("设置项目失败: %w", err)
	//}

	// 设置付款公司
	sendProgress(progress, "设置业务发生部门: "+payItem.PayDept)
	if err := handlePayDept(page, progress, payItem.PayDept); err != nil {
		sendProgress(progress, fmt.Sprintf("设置付款部门失败: %v", err))
		return fmt.Errorf("设置付款部门失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "支付信息填写完成")
	return nil
}

func handleBusinessDept(page playwright.Page, progress chan string, businessDeptText string) error {
	sendProgress(progress, "设置业务发生部门...")
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
	sendProgress(progress, "设置业务发生部门完成")
	return nil
}

func handleBudgetDept(page playwright.Page, progress chan string, budgetDept string) error {
	sendProgress(progress, "设置预算承担部门...")
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
	sendProgress(progress, "设置预算承担部门完成")
	return nil
}

func handleProjectType(page playwright.Page, progress chan string, projectType string) error {
	sendProgress(progress, "设置项目类型...")
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
	sendProgress(progress, "设置项目类型完成")
	return nil
}

func handleProject(page playwright.Page, progress chan string, project string) error {
	sendProgress(progress, "设置项目...")
	dialog := page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择项目\"]").Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击项目下拉框失败: %v", err))
		return fmt.Errorf("点击项目下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("获取下拉选项数量失败: %v", err))
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
					sendProgress(progress, fmt.Sprintf("选择项目失败: %v", err))
					return fmt.Errorf("选择项目失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "设置项目完成")
	return nil
}

func handlePayDept(page playwright.Page, progress chan string, payDept string) error {
	sendProgress(progress, "设置付款公司...")
	payLocator := page.Locator("span:has-text(\"付款公司\")").Locator("..").Locator("..").Locator(".el-input__inner")
	if err := payLocator.Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击付款公司下拉框失败: %v", err))
		return fmt.Errorf("点击付款公司下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("获取下拉选项数量失败: %v", err))
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
					sendProgress(progress, fmt.Sprintf("选择付款公司失败: %v", err))
					return fmt.Errorf("选择付款公司失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "设置付款部门完成")
	return nil
}

func handleAddDetail(page playwright.Page, progress chan string) error {
	sendProgress(progress, "新增报销明细...")
	addButton := page.Locator("button:has-text(\"导出\") + button")
	if err := addButton.Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击新增明细按钮失败: %v", err))
		return fmt.Errorf("点击新增明细按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "新增一条报销细节记录完成")
	return nil
}

func handleReimburseDetail(page playwright.Page, progress chan string, item CostItem, trIndex int) error {
	sendProgress(progress, fmt.Sprintf("填写报销明细第 %d 行...", trIndex))
	divElements := page.Locator("div.el-table__header-wrapper")
	count, err := divElements.Count()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("获取表格头数量失败: %v", err))
		return fmt.Errorf("获取表格头数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		divElement := divElements.Nth(i)
		if textCount, _ := divElement.GetByText("费用名称").Count(); textCount > 0 {
			trElements := divElement.Locator("+ div").Locator("tr")
			trCount, err := trElements.Count()
			if err != nil {
				sendProgress(progress, fmt.Sprintf("获取表格行数量失败: %v", err))
				return fmt.Errorf("获取表格行数量失败: %w", err)
			}

			for j := 0; j < trCount; j++ {
				if j+1 == trIndex {
					tdElements := trElements.Nth(j).Locator("td")
					tdCount, err := tdElements.Count()
					if err != nil {
						sendProgress(progress, fmt.Sprintf("获取表格列数量失败: %v", err))
						return fmt.Errorf("获取表格列数量失败: %w", err)
					}

					for k := 0; k < tdCount; k++ {
						tdElement := tdElements.Nth(k)
						switch k {
						case 1: // 费用类别
							if err := handleCostCategoryInDetail(page, progress, tdElement, item.Category); err != nil {
								sendProgress(progress, fmt.Sprintf("设置费用类别失败: %v", err))
								return fmt.Errorf("设置费用类别失败: %w", err)
							}
						case 2: // 费用名称
							if err := handleCostNameInDetail(page, progress, tdElement, item.Name); err != nil {
								sendProgress(progress, fmt.Sprintf("设置费用名称失败: %v", err))
								return fmt.Errorf("设置费用名称失败: %w", err)
							}
						case 3: // 费用说明
							if err := handleCostCommentInDetail(page, progress, tdElement, item.Comment); err != nil {
								sendProgress(progress, fmt.Sprintf("设置费用说明失败: %v", err))
								return fmt.Errorf("设置费用说明失败: %w", err)
							}
						case 4: // 报销金额
							if err := handleCostInDetail(page, progress, tdElement, item.Cost); err != nil {
								sendProgress(progress, fmt.Sprintf("设置报销金额失败: %v", err))
								return fmt.Errorf("设置报销金额失败: %w", err)
							}
						case 5: // 发票张数
							if err := handleBillNumberInDetail(page, progress, tdElement, item.BillNumber); err != nil {
								sendProgress(progress, fmt.Sprintf("设置发票张数失败: %v", err))
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
	sendProgress(progress, fmt.Sprintf("报销明细第 %d 行填写完成", trIndex))
	return nil
}

func handleCostCategoryInDetail(page playwright.Page, progress chan string, tdElement playwright.Locator, costCategory string) error {
	sendProgress(progress, "设置费用类别: "+costCategory)
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击费用类别输入框失败: %v", err))
		return fmt.Errorf("点击费用类别输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allCategoryItems := page.GetByText(costCategory)
	count, err := allCategoryItems.Count()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("获取费用类别选项数量失败: %v", err))
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
				sendProgress(progress, fmt.Sprintf("选择费用类别失败: %v", err))
				return fmt.Errorf("选择费用类别失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "费用类别设置完成")
	return nil
}

func handleCostNameInDetail(page playwright.Page, progress chan string, tdElement playwright.Locator, costName string) error {
	sendProgress(progress, "设置费用名称: "+costName)
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击费用名称输入框失败: %v", err))
		return fmt.Errorf("点击费用名称输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allNameItems := page.GetByText(costName)
	count, err := allNameItems.Count()
	if err != nil {
		sendProgress(progress, fmt.Sprintf("获取费用名称选项数量失败: %v", err))
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
				sendProgress(progress, fmt.Sprintf("选择费用名称失败: %v", err))
				return fmt.Errorf("选择费用名称失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "费用名称设置完成")
	return nil
}

func handleCostCommentInDetail(page playwright.Page, progress chan string, tdElement playwright.Locator, costComment string) error {
	sendProgress(progress, "填写费用说明: "+costComment)
	if err := tdElement.Locator("input.el-input__inner").Fill(costComment); err != nil {
		sendProgress(progress, fmt.Sprintf("填写费用说明失败: %v", err))
		return fmt.Errorf("填写费用说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "费用说明填写完成")
	return nil
}

func handleCostInDetail(page playwright.Page, progress chan string, tdElement playwright.Locator, cost string) error {
	sendProgress(progress, "填写报销金额: "+cost)
	if err := tdElement.Locator("input.el-input__inner").Fill(cost); err != nil {
		sendProgress(progress, fmt.Sprintf("填写报销金额失败: %v", err))
		return fmt.Errorf("填写报销金额失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "报销金额填写完成")
	return nil
}

func handleBillNumberInDetail(page playwright.Page, progress chan string, tdElement playwright.Locator, billNumber string) error {
	sendProgress(progress, "填写发票张数: "+billNumber)
	if err := tdElement.Locator("input.el-input__inner").Fill(billNumber); err != nil {
		sendProgress(progress, fmt.Sprintf("填写发票张数失败: %v", err))
		return fmt.Errorf("填写发票张数失败: %w", err)
	}

	time.Sleep(DelayShort)
	sendProgress(progress, "发票张数填写完成")
	return nil
}

func handleVatInvoiceUpload(page playwright.Page, progress chan string, fileName string) error {
	sendProgress(progress, "上传发票文件: "+fileName)
	if err := page.Locator("input.el-upload__input").SetInputFiles([]string{fileName}); err != nil {
		sendProgress(progress, fmt.Sprintf("上传发票文件失败: %v", err))
		return fmt.Errorf("上传发票文件失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	sendProgress(progress, "发票文件上传完成: "+fileName)
	return nil
}

func handleSaveInfo(page playwright.Page, progress chan string) error {
	sendProgress(progress, "保存信息...")
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "保存"}).Click(); err != nil {
		sendProgress(progress, fmt.Sprintf("点击保存按钮失败: %v", err))
		return fmt.Errorf("点击保存按钮失败: %w", err)
	}

	time.Sleep(DelayNormal)
	sendProgress(progress, "信息保存完成")
	return nil
}
