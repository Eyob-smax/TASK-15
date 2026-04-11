package http

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/security"
)

// RegisterRoutes configures all API route groups on the given Echo instance.
func RegisterRoutes(
	e *echo.Echo,
	logger *slog.Logger,
	pool *pgxpool.Pool,
	authHandler *AuthHandler,
	authMW echo.MiddlewareFunc,
	itemH *ItemHandler,
	inventoryH *InventoryHandler,
	campaignH *CampaignHandler,
	orderH *OrderHandler,
	supplierH *SupplierHandler,
	procurementH *ProcurementHandler,
	varianceH *VarianceHandler,
	reportH *ReportHandler,
	backupH *BackupHandler,
	retentionH *RetentionHandler,
	biometricH *BiometricHandler,
	userH *UserHandler,
	adminH *AdminHandler,
	dashboardH *DashboardHandler,
	locationH *LocationHandler,
	memberH *MemberHandler,
	coachH *CoachHandler,
	landedCostH *LandedCostHandler,
) {
	api := e.Group("/api/v1")

	api.Use(RequestIDMiddleware())
	api.Use(RecoverMiddleware(logger))

	registerAuthRoutes(api.Group("/auth"), authHandler, authMW)
	registerDashboardRoutes(api.Group("/dashboard"), authMW, dashboardH)
	registerItemRoutes(api.Group("/items"), authMW, itemH)
	registerInventoryRoutes(api.Group("/inventory"), authMW, inventoryH)
	registerWarehouseBinRoutes(api.Group("/warehouse-bins"), authMW, inventoryH)
	registerCampaignRoutes(api.Group("/campaigns"), authMW, campaignH)
	registerOrderRoutes(api.Group("/orders"), authMW, orderH)
	registerSupplierRoutes(api.Group("/suppliers"), authMW, supplierH)
	registerPurchaseOrderRoutes(api.Group("/purchase-orders"), authMW, procurementH)
	registerVarianceRoutes(api.Group("/variances"), authMW, varianceH)
	registerProcurementRoutes(api.Group("/procurement"), authMW, landedCostH)
	registerReportRoutes(api.Group("/reports"), authMW, reportH)
	registerExportRoutes(api.Group("/exports"), authMW, reportH)
	registerAdminRoutes(api.Group("/admin"), authMW, backupH, retentionH, biometricH, userH, adminH)
	registerLocationRoutes(api.Group("/locations"), authMW, locationH)
	registerCoachRoutes(api.Group("/coaches"), authMW, coachH)
	registerMemberRoutes(api.Group("/members"), authMW, memberH)
}

func registerAuthRoutes(g *echo.Group, authHandler *AuthHandler, authMW echo.MiddlewareFunc) {
	g.POST("/login", authHandler.Login)
	g.POST("/logout", authHandler.Logout, authMW)
	g.GET("/session", authHandler.GetSession, authMW)
	g.POST("/captcha/verify", authHandler.VerifyCaptcha)
}

func registerDashboardRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *DashboardHandler) {
	viewDashboard := NewRequireRole(security.ActionViewDashboard)
	g.Use(authMW)
	g.GET("/kpis", h.GetKPIs, viewDashboard)
}

func registerItemRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *ItemHandler) {
	manageCatalog := NewRequireRole(security.ActionManageCatalog)
	viewCatalog := NewRequireRole(security.ActionViewCatalog, security.ActionManageCatalog)

	g.Use(authMW)
	g.POST("", h.CreateItem, manageCatalog)
	g.GET("", h.ListItems, viewCatalog)
	g.GET("/:id", h.GetItem, viewCatalog)
	g.PUT("/:id", h.UpdateItem, manageCatalog)
	g.POST("/:id/publish", h.PublishItem, manageCatalog)
	g.POST("/:id/unpublish", h.UnpublishItem, manageCatalog)
	g.POST("/batch-edit", h.BatchEdit, manageCatalog)
}

func registerInventoryRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *InventoryHandler) {
	manageInventory := NewRequireRole(security.ActionManageInventory)
	// Inventory snapshots are operational data — restrict to roles that actively
	// manage stock or procurement (Admin, OpsMgr, ProcurementSpecialist).
	// ViewCatalog is intentionally excluded: it is granted to all roles including
	// Coach and Member, who must not see raw stock-level data.
	viewInventory := NewRequireRole(security.ActionManageInventory, security.ActionManageProcurement)

	g.Use(authMW)
	g.GET("/snapshots", h.GetSnapshots, viewInventory)
	g.POST("/adjustments", h.CreateAdjustment, manageInventory)
	g.GET("/adjustments", h.ListAdjustments, manageInventory)
}

func registerWarehouseBinRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *InventoryHandler) {
	manageInventory := NewRequireRole(security.ActionManageInventory)

	g.Use(authMW)
	g.POST("", h.CreateWarehouseBin, manageInventory)
	g.GET("", h.ListWarehouseBins, manageInventory)
	g.GET("/:id", h.GetWarehouseBin, manageInventory)
}

func registerCampaignRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *CampaignHandler) {
	createCampaign := NewRequireRole(security.ActionCreateCampaign)
	viewCatalog := NewRequireRole(security.ActionViewCatalog, security.ActionManageCatalog)
	joinCampaign := NewRequireRole(security.ActionJoinCampaign)
	manageCatalog := NewRequireRole(security.ActionManageCatalog)

	g.Use(authMW)
	g.POST("", h.CreateCampaign, createCampaign)
	g.GET("", h.ListCampaigns, viewCatalog)
	g.GET("/:id", h.GetCampaign, viewCatalog)
	g.POST("/:id/join", h.JoinCampaign, joinCampaign)
	g.POST("/:id/cancel", h.CancelCampaign, manageCatalog)
	g.POST("/:id/evaluate", h.EvaluateCampaign, manageCatalog)
}

func registerOrderRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *OrderHandler) {
	manageOrders := NewRequireRole(security.ActionManageOrders)
	viewOrders := NewRequireRole(security.ActionViewOrders, security.ActionManageOrders)
	createOrder := NewRequireRole(security.ActionCreateOrder, security.ActionManageOrders)

	g.Use(authMW)
	g.POST("", h.CreateOrder, createOrder)
	g.GET("", h.ListOrders, viewOrders)
	g.GET("/:id", h.GetOrder, viewOrders)
	g.POST("/:id/pay", h.PayOrder, manageOrders)
	// POST /orders/:id/cancel — authMW only; ownership+status check in handler.
	g.POST("/:id/cancel", h.CancelOrder)
	g.POST("/:id/refund", h.RefundOrder, manageOrders)
	g.POST("/:id/notes", h.AddOrderNote, manageOrders)
	g.GET("/:id/timeline", h.GetOrderTimeline, viewOrders)
	g.POST("/:id/split", h.SplitOrder, manageOrders)
	g.POST("/merge", h.MergeOrders, manageOrders)
}

func registerSupplierRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *SupplierHandler) {
	manageProcurement := NewRequireRole(security.ActionManageProcurement)
	viewProcurement := NewRequireRole(security.ActionViewProcurement, security.ActionManageProcurement)
	g.Use(authMW)
	g.POST("", h.CreateSupplier, manageProcurement)
	g.GET("", h.ListSuppliers, viewProcurement)
	g.GET("/:id", h.GetSupplier, viewProcurement)
	g.PUT("/:id", h.UpdateSupplier, manageProcurement)
}

func registerPurchaseOrderRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *ProcurementHandler) {
	manageProcurement := NewRequireRole(security.ActionManageProcurement)
	viewProcurement := NewRequireRole(security.ActionViewProcurement, security.ActionManageProcurement)
	g.Use(authMW)
	g.POST("", h.CreatePO, manageProcurement)
	g.GET("", h.ListPOs, viewProcurement)
	g.GET("/:id", h.GetPO, viewProcurement)
	g.POST("/:id/approve", h.ApprovePO, manageProcurement)
	g.POST("/:id/receive", h.ReceivePO, manageProcurement)
	g.POST("/:id/return", h.ReturnPO, manageProcurement)
	g.POST("/:id/void", h.VoidPO, manageProcurement)
}

func registerVarianceRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *VarianceHandler) {
	manageProcurement := NewRequireRole(security.ActionManageProcurement)
	viewProcurement := NewRequireRole(security.ActionViewProcurement, security.ActionManageProcurement)
	g.Use(authMW)
	g.GET("", h.ListVariances, viewProcurement)
	g.GET("/:id", h.GetVariance, viewProcurement)
	g.POST("/:id/resolve", h.ResolveVariance, manageProcurement)
}

func registerProcurementRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *LandedCostHandler) {
	viewProcurement := NewRequireRole(security.ActionViewProcurement, security.ActionManageProcurement)
	g.Use(authMW)
	g.GET("/landed-costs", h.GetLandedCosts, viewProcurement)
	g.GET("/landed-costs/:poId", h.GetLandedCostsByPO, viewProcurement)
}

func registerReportRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *ReportHandler) {
	viewReports := NewRequireRole(security.ActionViewReports, security.ActionManageReports)
	g.Use(authMW)
	g.GET("", h.ListReports, viewReports)
	g.GET("/:id/data", h.GetReportData, viewReports)
}

func registerExportRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *ReportHandler) {
	viewReports := NewRequireRole(security.ActionViewReports, security.ActionManageReports)
	g.Use(authMW)
	g.POST("", h.RunExport, viewReports)
	g.GET("/:id", h.GetExport, viewReports)
	g.GET("/:id/download", h.DownloadExport, viewReports)
}

func registerAdminRoutes(g *echo.Group, authMW echo.MiddlewareFunc, backupH *BackupHandler, retentionH *RetentionHandler, biometricH *BiometricHandler, userH *UserHandler, adminH *AdminHandler) {
	viewAudit := NewRequireRole(security.ActionViewAudit)
	manageUsers := NewRequireRole(security.ActionManageUsers)
	manageBackups := NewRequireRole(security.ActionManageBackups)
	manageBiometric := NewRequireRole(security.ActionManageBiometric)

	g.Use(authMW)

	// Audit
	g.GET("/audit-log", adminH.GetAuditLog, viewAudit)
	g.GET("/audit-log/security", adminH.GetSecurityEvents, viewAudit)

	// Backups
	g.POST("/backups", backupH.TriggerBackup, manageBackups)
	g.GET("/backups", backupH.ListBackups, manageBackups)

	// Biometrics — static routes registered before param routes to ensure correct matching
	g.POST("/biometrics/rotate-key", biometricH.RotateKey, manageBiometric)
	g.GET("/biometrics/keys", biometricH.ListKeys, manageBiometric)
	g.POST("/biometrics", biometricH.RegisterBiometric, manageBiometric)
	g.GET("/biometrics/:user_id", biometricH.GetBiometric, manageBiometric)
	g.POST("/biometrics/:user_id/revoke", biometricH.RevokeBiometric, manageBiometric)

	// Users
	g.POST("/users", userH.CreateUser, manageUsers)
	g.GET("/users", userH.ListUsers, manageUsers)
	g.GET("/users/:id", userH.GetUser, manageUsers)
	g.PUT("/users/:id", userH.UpdateUser, manageUsers)
	g.POST("/users/:id/deactivate", userH.DeactivateUser, manageUsers)

	// Retention policies
	g.GET("/retention-policies", retentionH.ListRetentionPolicies, manageBackups)
	g.GET("/retention-policies/:entity_type", retentionH.GetRetentionPolicy, manageBackups)
	g.PUT("/retention-policies/:entity_type", retentionH.UpdateRetentionPolicy, manageBackups)
}

func registerLocationRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *LocationHandler) {
	manageLocations := NewRequireRole(security.ActionManageInventory)
	// ActionViewLocations is intentionally distinct from ActionViewCatalog so that
	// Member role (which has ViewCatalog) cannot access operational location data.
	viewLocations := NewRequireRole(security.ActionViewLocations, security.ActionManageInventory)
	g.Use(authMW)
	g.POST("", h.CreateLocation, manageLocations)
	g.GET("", h.ListLocations, viewLocations)
	g.GET("/:id", h.GetLocation, viewLocations)
}

func registerCoachRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *CoachHandler) {
	manageUsers := NewRequireRole(security.ActionManageUsers)
	// ActionViewCoaches is separate from ActionViewCatalog so Members cannot
	// enumerate personnel records.
	viewCoaches := NewRequireRole(security.ActionViewCoaches, security.ActionManageUsers)
	g.Use(authMW)
	g.POST("", h.CreateCoach, manageUsers)
	g.GET("", h.ListCoaches, viewCoaches)
	g.GET("/:id", h.GetCoach, viewCoaches)
}

func registerMemberRoutes(g *echo.Group, authMW echo.MiddlewareFunc, h *MemberHandler) {
	manageUsers := NewRequireRole(security.ActionManageUsers)
	// ActionViewMembers is separate from ActionViewCatalog so regular Members cannot
	// enumerate other members' records.
	viewMembers := NewRequireRole(security.ActionViewMembers, security.ActionManageUsers)
	g.Use(authMW)
	g.POST("", h.CreateMember, manageUsers)
	g.GET("", h.ListMembers, viewMembers)
	g.GET("/:id", h.GetMember, viewMembers)
}
