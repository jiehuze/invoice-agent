package services

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"invoice-agent/internal/app/models"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// 常量定义
const (
	DelayShort  = 1 * time.Second
	DelayNormal = 10 * time.Second
	DelayLong   = 15 * time.Second
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// 任务信息
type TaskInfo struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Progress  string     `json:"progress"`
	Error     string     `json:"error,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	// 进度通道
	progressChan chan string
	// 取消信号
	cancelChan chan struct{}
	// 完成信号
	doneChan chan struct{}
}

// 自动化填报实例
type AutoFillingInstance struct {
	taskID  string
	request *models.AutoFillingRequest
	//progress chan string
	cancel  chan struct{}
	status  TaskStatus
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
	mutex   sync.RWMutex
}

// 自动化填报服务（多实例管理器）
type AutoFillingService struct {
	tasks     sync.Map // taskID -> *TaskInfo
	instances sync.Map // taskID -> *AutoFillingInstance

	// Playwright 安装状态
	playwrightInstalled bool
	installMutex        sync.Mutex
}

// 创建新的服务实例
func NewAutoFillingService() *AutoFillingService {
	return &AutoFillingService{
		tasks:     sync.Map{},
		instances: sync.Map{},
	}
}

// 发送进度信息
func (s *AutoFillingService) sendProgress(taskInfo *TaskInfo, message string) {
	taskInfo.Progress = message
	s.tasks.Store(taskInfo.ID, taskInfo)

	select {
	case taskInfo.progressChan <- message + "\n":
		// 进度信息已发送
	default:
		// 通道已满，跳过
	}
	log.Infof("[Task %s] %s", taskInfo.ID, message)
}

// 创建新的自动化填报实例
func (s *AutoFillingService) NewAutoFillingInstance(taskID string, req *models.AutoFillingRequest) *AutoFillingInstance {
	return &AutoFillingInstance{
		taskID:  taskID,
		request: req,
		cancel:  make(chan struct{}),
		status:  TaskStatusPending,
	}
}

// 开始自动化填报任务
func (s *AutoFillingService) StartAutoFilling(taskID string, req *models.AutoFillingRequest) error {
	// 检查任务是否已存在
	if _, exists := s.tasks.Load(taskID); exists {
		return fmt.Errorf("任务ID已存在: %s", taskID)
	}

	// 创建任务信息
	taskInfo := &TaskInfo{
		ID:           taskID,
		Status:       TaskStatusPending,
		Progress:     "任务初始化中...",
		StartedAt:    time.Now(),
		progressChan: make(chan string),
		cancelChan:   make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
	s.tasks.Store(taskID, taskInfo)

	// 创建实例
	instance := s.NewAutoFillingInstance(taskID, req)
	s.instances.Store(taskID, instance)

	// 异步执行任务
	go s.executeTask(taskID, instance, taskInfo)

	return nil
}

// 执行任务
func (s *AutoFillingService) executeTask(taskID string, instance *AutoFillingInstance, taskInfo *TaskInfo) {
	// 确保在函数退出时关闭进度通道
	defer func() {
		// 先关闭进度通道，让外部监听循环能够退出
		close(taskInfo.progressChan)
		close(taskInfo.doneChan)

		if r := recover(); r != nil {
			taskInfo.Status = TaskStatusFailed
			taskInfo.Error = fmt.Sprintf("任务执行异常: %v", r)
			now := time.Now()
			taskInfo.EndedAt = &now
			s.tasks.Store(taskID, taskInfo)
		}

		// 清理实例资源
		s.cleanupInstance(instance)
	}()

	// 更新任务状态
	taskInfo.Status = TaskStatusRunning
	taskInfo.Progress = "开始执行自动化填报..."
	s.tasks.Store(taskID, taskInfo)

	// 执行填报流程
	err := s.runAutoFilling(instance, taskInfo)
	if err != nil {
		taskInfo.Status = TaskStatusFailed
		taskInfo.Error = err.Error()
	} else {
		taskInfo.Status = TaskStatusCompleted
		taskInfo.Progress = "自动化填报完成"
	}

	now := time.Now()
	taskInfo.EndedAt = &now
	s.tasks.Store(taskID, taskInfo)
}

// 清理实例资源
func (s *AutoFillingService) cleanupInstance(instance *AutoFillingInstance) {
	if instance == nil {
		return
	}

	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.page != nil {
		instance.page.Close()
	}
	if instance.context != nil {
		instance.context.Close()
	}
	if instance.browser != nil {
		instance.browser.Close()
	}
	if instance.pw != nil {
		instance.pw.Stop()
	}

	//close(instance.progress)
}

// 核心执行逻辑
func (s *AutoFillingService) runAutoFilling(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	// 确保 Playwright 已安装
	if err := s.ensurePlaywrightInstalled(); err != nil {
		return fmt.Errorf("Playwright 初始化失败: %w", err)
	}

	s.sendProgress(taskInfo, "开始初始化 Playwright...")

	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("启动 Playwright 失败: %w", err)
	}
	instance.pw = pw

	s.sendProgress(taskInfo, "浏览器启动中...")

	// 启动浏览器
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("启动浏览器失败: %w", err)
	}
	instance.browser = browser

	s.sendProgress(taskInfo, "浏览器启动成功")

	// 创建浏览器上下文
	context, err := browser.NewContext()
	if err != nil {
		return fmt.Errorf("创建浏览器上下文失败: %w", err)
	}
	instance.context = context

	s.sendProgress(taskInfo, "浏览器上下文创建成功")

	// 创建页面
	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("创建页面失败: %w", err)
	}
	instance.page = page

	s.sendProgress(taskInfo, "新页面创建成功")

	// 执行主要流程
	return s.executeFillingProcess(instance, taskInfo)
}

// 执行填报流程
func (s *AutoFillingService) executeFillingProcess(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	// 导航到目标URL
	s.sendProgress(taskInfo, "正在导航到目标网址...")
	if _, err := instance.page.Goto("http://open.sky-dome.com.cn:9086/"); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}

	// 检查是否被取消
	if s.isTaskCancelled(instance) {
		return fmt.Errorf("任务已被取消")
	}

	// 执行登录操作
	s.sendProgress(taskInfo, "开始执行登录操作...")
	if err := s.handleLogin(instance, taskInfo); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	// 后续流程与之前类似，但需要传递 instance 和 taskInfo
	// 打开新增对话框
	s.sendProgress(taskInfo, "正在打开新增对话框...")
	if err := s.handleAddDialog(instance, taskInfo); err != nil {
		return fmt.Errorf("打开新增对话框失败: %w", err)
	}

	// 填写基础信息
	s.sendProgress(taskInfo, "填写基础信息...")
	if err := s.handleReimburseBasic(instance, taskInfo); err != nil {
		return fmt.Errorf("填写基础信息失败: %w", err)
	}

	if err := s.handleWheelDown(instance, 0, 500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	// 填写支付信息
	s.sendProgress(taskInfo, "填写支付信息...")
	if err := s.handleReimbursePayInfo(instance, taskInfo); err != nil {
		return fmt.Errorf("填写支付信息失败: %w", err)
	}

	// 填写报销明细
	s.sendProgress(taskInfo, "填写报销明细...")
	for i := 0; i < len(instance.request.CostItems); i++ {
		if err := s.handleAddDetail(instance, taskInfo); err != nil {
			return fmt.Errorf("添加明细失败: %w", err)
		}
	}

	for i, item := range instance.request.CostItems {
		if err := s.handleReimburseDetail(instance, taskInfo, item, i+1); err != nil {
			return fmt.Errorf("填写报销明细失败: %w", err)
		}
	}

	// 增值税发票批量上传
	if err := s.handleWheelDown(instance, 0, 500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}
	s.sendProgress(taskInfo, "上传发票...")
	for _, filePath := range instance.request.FilePaths {
		if err := s.handleVatInvoiceUpload(instance, taskInfo, filePath); err != nil {
			return fmt.Errorf("上传发票失败: %w", err)
		}
	}

	if err := s.handleWheelDown(instance, 0, -500); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	if err := s.handleSaveInfo(instance, taskInfo); err != nil {
		return fmt.Errorf("保存信息失败: %w", err)
	}

	return nil
}

// 检查任务是否被取消
func (s *AutoFillingService) isTaskCancelled(instance *AutoFillingInstance) bool {
	select {
	case <-instance.cancel:
		return true
	default:
		return false
	}
}

// 确保 Playwright 已安装
func (s *AutoFillingService) ensurePlaywrightInstalled() error {
	s.installMutex.Lock()
	defer s.installMutex.Unlock()

	if !s.playwrightInstalled {
		if err := playwright.Install(); err != nil {
			return fmt.Errorf("安装 Playwright 失败: %w", err)
		}
		s.playwrightInstalled = true
	}
	return nil
}

// 获取任务状态
func (s *AutoFillingService) GetTaskStatus(taskID string) (*TaskInfo, bool) {
	status, exists := s.tasks.Load(taskID)
	if !exists {
		return nil, false
	}
	return status.(*TaskInfo), true
}

// 获取任务进度通道
func (s *AutoFillingService) GetTaskProgressChan(taskID string) (chan string, bool) {
	taskInfo, exists := s.tasks.Load(taskID)
	if !exists {
		return nil, false
	}
	return taskInfo.(*TaskInfo).progressChan, true
}

// 取消任务
func (s *AutoFillingService) CancelTask(taskID string) bool {
	taskInfo, exists := s.tasks.Load(taskID)
	if !exists {
		return false
	}

	info := taskInfo.(*TaskInfo)
	if info.Status == TaskStatusRunning {
		info.Status = TaskStatusCancelled
		s.tasks.Store(taskID, info)

		// 发送取消信号到实例
		if instance, exists := s.instances.Load(taskID); exists {
			close(instance.(*AutoFillingInstance).cancel)
		}
		// 发送取消进度信息
		s.sendProgress(info, "任务已被取消")

		// 关闭进度通道
		close(info.progressChan)
		return true
	}
	return false
}

// 清理已完成的任务
func (s *AutoFillingService) CleanupCompletedTasks(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	s.tasks.Range(func(key, value interface{}) bool {
		taskInfo := value.(*TaskInfo)
		if taskInfo.EndedAt != nil && taskInfo.EndedAt.Before(cutoff) {
			s.tasks.Delete(key)
			s.instances.Delete(key)
		}
		return true
	})
}

var AutoFillingClient *AutoFillingService

//func NewAutoFillingService() *AutoFillingService {
//	if AutoFillingClient == nil {
//		AutoFillingClient = &AutoFillingService{}
//	}
//	return AutoFillingClient
//}

func (s *AutoFillingService) handleLogin(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "进入登录页面...")

	// 等待用户名输入框
	s.sendProgress(taskInfo, "等待输入用户名和密码...")
	if err := instance.page.Locator("[placeholder=\"请输入账号\"]").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("等待用户名输入框失败:  %v", err))
		return fmt.Errorf("等待用户名输入框失败: %w", err)
	}

	// 填写用户名
	if err := instance.page.Locator("[placeholder=\"请输入账号\"]").Fill("tyzq-wangmeng5"); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写用户名失败:  %v", err))
		return fmt.Errorf("填写用户名失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 填写密码
	if err := instance.page.Locator("[placeholder=\"请输入密码\"]").Fill("tyzq123456"); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写密码失败:  %v", err))
		return fmt.Errorf("填写密码失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 点击登录按钮
	s.sendProgress(taskInfo, "登录系统...")
	if err := instance.page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "登录"}).Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击登录按钮失败:  %v", err))
		return fmt.Errorf("点击登录按钮失败: %w", err)
	}

	// 等待并重新加载
	time.Sleep(DelayLong)
	s.sendProgress(taskInfo, "导航到报销页面...")
	if _, err := instance.page.Goto("http://open.sky-dome.com.cn:9086/#/reimbursement/employee"); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("导航到报销页面失败:  %v", err))
		return fmt.Errorf("导航到报销页面失败: %w", err)
	}

	time.Sleep(DelayNormal)
	s.sendProgress(taskInfo, "进入到报销页面...")
	return nil
}

func (s *AutoFillingService) handleAddDialog(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "打开新增报销记录对话框...")
	if err := instance.page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "新增"}).Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击新增按钮失败:  %v", err))
		return fmt.Errorf("点击新增按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "打开新增报销记录对话框完成")
	return nil
}

func (s *AutoFillingService) handleReimburseBasic(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	// 设置报销类型
	s.sendProgress(taskInfo, "设置报销类型...")
	if err := s.handleReimburseType(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置报销类型失败:  %v", err))
		return fmt.Errorf("设置报销类型失败: %w", err)
	}

	// 设置紧急类型
	s.sendProgress(taskInfo, "设置紧急类型...")
	if err := s.handleUrgentType(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置紧急类型失败: %v", err))
		return fmt.Errorf("设置紧急类型失败: %w", err)
	}

	// 填写报销说明
	s.sendProgress(taskInfo, "填写报销说明...")
	if err := s.handleReimburseComment(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写报销说明失败: %v", err))
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "基础信息填写完成...")
	return nil
}

func (s *AutoFillingService) handleReimburseType(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "填写报销类型...")
	dialog := instance.page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择报销类型\"]").Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击报销类型下拉框失败: %v", err))
		return fmt.Errorf("点击报销类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择报销类型
	if err := instance.page.Locator(fmt.Sprintf("li.el-select-dropdown__item:has-text(\"%s\")", instance.request.BasicInfo.Category)).Last().Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("选择报销类型失败: %v", err))
		return fmt.Errorf("选择报销类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置报销类型完成...")
	return nil
}

func (s *AutoFillingService) handleUrgentType(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "设置紧急类型...")
	dialog := instance.page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择紧急类型\"]").Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击紧急类型下拉框失败: %v", err))
		return fmt.Errorf("点击紧急类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	// 选择紧急类型
	if err := instance.page.GetByText(instance.request.BasicInfo.UrgentType, playwright.PageGetByTextOptions{Exact: playwright.Bool(true)}).Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("选择紧急类型失败: %v", err))
		return fmt.Errorf("选择紧急类型失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置紧急类型完成...")
	return nil
}

func (s *AutoFillingService) handleReimburseComment(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "填写报销说明:"+instance.request.BasicInfo.Comment)
	if err := instance.page.Locator("[placeholder=\"请输入报销说明\"]").Fill(instance.request.BasicInfo.Comment); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写报销说明失败: %v", err))
		return fmt.Errorf("填写报销说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置报销说明完成...")
	return nil
}

func (s *AutoFillingService) handleWheelDown(instance *AutoFillingInstance, x float64, y float64) error {
	if err := instance.page.Mouse().Wheel(x, y); err != nil {
		return fmt.Errorf("滚动失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	return nil
}

func (s *AutoFillingService) handleReimbursePayInfo(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	// 设置业务发生部门
	s.sendProgress(taskInfo, "设置业务发生部门: "+instance.request.PayInfo.BusinessDept)
	if err := s.handleBusinessDept(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置业务发生部门失败: %v", err))
		return fmt.Errorf("设置业务发生部门失败: %w", err)
	}

	// 设置预算承担部门
	s.sendProgress(taskInfo, "设置预算承担部门: "+instance.request.PayInfo.BudgetDept)
	if err := s.handleBudgetDept(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置预算承担部门失败: %v", err))
		return fmt.Errorf("设置预算承担部门失败: %w", err)
	}

	// 设置费用归集项目
	s.sendProgress(taskInfo, "设置费用归集项目: "+instance.request.PayInfo.ProjectType)
	if err := s.handleProjectType(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置项目类型失败: %v", err))
		return fmt.Errorf("设置项目类型失败: %w", err)
	}
	//if err := handleProject(page, payItem.Project); err != nil {
	//	return fmt.Errorf("设置项目失败: %w", err)
	//}

	// 设置付款公司
	s.sendProgress(taskInfo, "设置付款公司: "+instance.request.PayInfo.PayDept)
	if err := s.handlePayDept(instance, taskInfo); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("设置付款部门失败: %v", err))
		return fmt.Errorf("设置付款部门失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "支付信息填写完成")
	return nil
}

func (s *AutoFillingService) handleBusinessDept(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "设置业务发生部门...")
	businessDept := instance.page.Locator("div.el-form-item.is-required.custom-form-render-item.custom-form-render-item-twoline")
	if err := businessDept.Locator("[placeholder='请选择']").Click(); err != nil {
		return fmt.Errorf("点击业务发生部门下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := instance.page.Locator(".el-select-dropdown__item")
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
			textLocator := liElement.GetByText(instance.request.PayInfo.BusinessDept)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择业务发生部门失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置业务发生部门完成")
	return nil
}

func (s *AutoFillingService) handleBudgetDept(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "设置预算承担部门...")
	budgetLocator := instance.page.Locator("span:has-text(\"预算承担部门\")").Locator("..").Locator("..").Locator(".el-input__inner")
	if err := budgetLocator.Click(); err != nil {
		return fmt.Errorf("点击预算承担部门下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := instance.page.Locator(".el-select-dropdown__item")
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
			textLocator := liElement.GetByText(instance.request.PayInfo.BudgetDept)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择预算承担部门失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置预算承担部门完成")
	return nil
}

func (s *AutoFillingService) handleProjectType(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "设置项目类型...")
	dialog := instance.page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"项目类型\"]").Click(); err != nil {
		return fmt.Errorf("点击项目类型下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := instance.page.Locator(".el-select-dropdown__item")
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
			textLocator := liElement.GetByText(instance.request.PayInfo.ProjectType)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					return fmt.Errorf("选择项目类型失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置项目类型完成")
	return nil
}

func (s *AutoFillingService) handleProject(instance *AutoFillingInstance, taskInfo *TaskInfo, project string) error {
	s.sendProgress(taskInfo, "设置项目...")
	dialog := instance.page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "dialog"})
	if err := dialog.Locator("[placeholder=\"请选择项目\"]").Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击项目下拉框失败: %v", err))
		return fmt.Errorf("点击项目下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := instance.page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("获取下拉选项数量失败: %v", err))
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
					s.sendProgress(taskInfo, fmt.Sprintf("选择项目失败: %v", err))
					return fmt.Errorf("选择项目失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置项目完成")
	return nil
}

func (s *AutoFillingService) handlePayDept(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "设置付款公司...")
	payLocator := instance.page.Locator("span:has-text(\"付款公司\")").Locator("..").Locator("..").Locator(".el-input__inner")
	if err := payLocator.Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击付款公司下拉框失败: %v", err))
		return fmt.Errorf("点击付款公司下拉框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allLiElements := instance.page.Locator(".el-select-dropdown__item")
	count, err := allLiElements.Count()
	if err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("获取下拉选项数量失败: %v", err))
		return fmt.Errorf("获取下拉选项数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		liElement := allLiElements.Nth(i)
		visible, err := liElement.IsVisible()
		if err != nil {
			continue
		}
		if visible {
			textLocator := liElement.GetByText(instance.request.PayInfo.PayDept)
			if textCount, _ := textLocator.Count(); textCount > 0 {
				if err := textLocator.Click(); err != nil {
					s.sendProgress(taskInfo, fmt.Sprintf("选择付款公司失败: %v", err))
					return fmt.Errorf("选择付款公司失败: %w", err)
				}
				break
			}
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "设置付款部门完成")
	return nil
}

func (s *AutoFillingService) handleAddDetail(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "新增报销明细...")
	addButton := instance.page.Locator("button:has-text(\"导出\") + button")
	if err := addButton.Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击新增明细按钮失败: %v", err))
		return fmt.Errorf("点击新增明细按钮失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "新增一条报销细节记录完成")
	return nil
}

func (s *AutoFillingService) handleReimburseDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, item models.CostItem, trIndex int) error {
	s.sendProgress(taskInfo, fmt.Sprintf("填写报销明细第 %d 行...", trIndex))
	divElements := instance.page.Locator("div.el-table__header-wrapper")
	count, err := divElements.Count()
	if err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("获取表格头数量失败: %v", err))
		return fmt.Errorf("获取表格头数量失败: %w", err)
	}

	for i := 0; i < count; i++ {
		divElement := divElements.Nth(i)
		if textCount, _ := divElement.GetByText("费用名称").Count(); textCount > 0 {
			trElements := divElement.Locator("+ div").Locator("tr")
			trCount, err := trElements.Count()
			if err != nil {
				s.sendProgress(taskInfo, fmt.Sprintf("获取表格行数量失败: %v", err))
				return fmt.Errorf("获取表格行数量失败: %w", err)
			}

			for j := 0; j < trCount; j++ {
				if j+1 == trIndex {
					tdElements := trElements.Nth(j).Locator("td")
					tdCount, err := tdElements.Count()
					if err != nil {
						s.sendProgress(taskInfo, fmt.Sprintf("获取表格列数量失败: %v", err))
						return fmt.Errorf("获取表格列数量失败: %w", err)
					}

					for k := 0; k < tdCount; k++ {
						tdElement := tdElements.Nth(k)
						switch k {
						case 1: // 费用类别
							if err := s.handleCostCategoryInDetail(instance, taskInfo, tdElement, item.Category); err != nil {
								s.sendProgress(taskInfo, fmt.Sprintf("设置费用类别失败: %v", err))
								return fmt.Errorf("设置费用类别失败: %w", err)
							}
						case 2: // 费用名称
							if err := s.handleCostNameInDetail(instance, taskInfo, tdElement, item.Name); err != nil {
								s.sendProgress(taskInfo, fmt.Sprintf("设置费用名称失败: %v", err))
								return fmt.Errorf("设置费用名称失败: %w", err)
							}
						case 3: // 费用说明
							if err := s.handleCostCommentInDetail(instance, taskInfo, tdElement, item.Comment); err != nil {
								s.sendProgress(taskInfo, fmt.Sprintf("设置费用说明失败: %v", err))
								return fmt.Errorf("设置费用说明失败: %w", err)
							}
						case 4: // 报销金额
							if err := s.handleCostInDetail(instance, taskInfo, tdElement, item.Cost); err != nil {
								s.sendProgress(taskInfo, fmt.Sprintf("设置报销金额失败: %v", err))
								return fmt.Errorf("设置报销金额失败: %w", err)
							}
						case 5: // 发票张数
							if err := s.handleBillNumberInDetail(instance, taskInfo, tdElement, item.BillNumber); err != nil {
								s.sendProgress(taskInfo, fmt.Sprintf("设置发票张数失败: %v", err))
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
	s.sendProgress(taskInfo, fmt.Sprintf("报销明细第 %d 行填写完成", trIndex))
	return nil
}

func (s *AutoFillingService) handleCostCategoryInDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, tdElement playwright.Locator, costCategory string) error {
	s.sendProgress(taskInfo, "设置费用类别: "+costCategory)
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击费用类别输入框失败: %v", err))
		return fmt.Errorf("点击费用类别输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allCategoryItems := instance.page.GetByText(costCategory)
	count, err := allCategoryItems.Count()
	if err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("获取费用类别选项数量失败: %v", err))
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
				s.sendProgress(taskInfo, fmt.Sprintf("选择费用类别失败: %v", err))
				return fmt.Errorf("选择费用类别失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "费用类别设置完成")
	return nil
}

func (s *AutoFillingService) handleCostNameInDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, tdElement playwright.Locator, costName string) error {
	s.sendProgress(taskInfo, "设置费用名称: "+costName)
	if err := tdElement.Locator("input.el-input__inner").Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击费用名称输入框失败: %v", err))
		return fmt.Errorf("点击费用名称输入框失败: %w", err)
	}

	time.Sleep(DelayShort)

	allNameItems := instance.page.GetByText(costName)
	count, err := allNameItems.Count()
	if err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("获取费用名称选项数量失败: %v", err))
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
				s.sendProgress(taskInfo, fmt.Sprintf("选择费用名称失败: %v", err))
				return fmt.Errorf("选择费用名称失败: %w", err)
			}
			break
		}
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "费用名称设置完成")
	return nil
}

func (s *AutoFillingService) handleCostCommentInDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, tdElement playwright.Locator, costComment string) error {
	s.sendProgress(taskInfo, "填写费用说明: "+costComment)
	if err := tdElement.Locator("input.el-input__inner").Fill(costComment); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写费用说明失败: %v", err))
		return fmt.Errorf("填写费用说明失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "费用说明填写完成")
	return nil
}

func (s *AutoFillingService) handleCostInDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, tdElement playwright.Locator, cost string) error {
	s.sendProgress(taskInfo, "填写报销金额: "+cost)
	if err := tdElement.Locator("input.el-input__inner").Fill(cost); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写报销金额失败: %v", err))
		return fmt.Errorf("填写报销金额失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "报销金额填写完成")
	return nil
}

func (s *AutoFillingService) handleBillNumberInDetail(instance *AutoFillingInstance, taskInfo *TaskInfo, tdElement playwright.Locator, billNumber string) error {
	s.sendProgress(taskInfo, "填写发票张数: "+billNumber)
	if err := tdElement.Locator("input.el-input__inner").Fill(billNumber); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("填写发票张数失败: %v", err))
		return fmt.Errorf("填写发票张数失败: %w", err)
	}

	time.Sleep(DelayShort)
	s.sendProgress(taskInfo, "发票张数填写完成")
	return nil
}

func (s *AutoFillingService) handleVatInvoiceUpload(instance *AutoFillingInstance, taskInfo *TaskInfo, fileName string) error {
	s.sendProgress(taskInfo, "上传发票文件: "+fileName)
	if err := instance.page.Locator("input.el-upload__input").SetInputFiles([]string{fileName}); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("上传发票文件失败: %v", err))
		return fmt.Errorf("上传发票文件失败: %w", err)
	}

	time.Sleep(5 * DelayShort)
	s.sendProgress(taskInfo, "发票文件上传完成: "+fileName)
	return nil
}

func (s *AutoFillingService) handleSaveInfo(instance *AutoFillingInstance, taskInfo *TaskInfo) error {
	s.sendProgress(taskInfo, "保存信息...")
	if err := instance.page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "保存"}).Click(); err != nil {
		s.sendProgress(taskInfo, fmt.Sprintf("点击保存按钮失败: %v", err))
		return fmt.Errorf("点击保存按钮失败: %w", err)
	}

	time.Sleep(DelayNormal)
	s.sendProgress(taskInfo, "信息保存完成")
	return nil
}
