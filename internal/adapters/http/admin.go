package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wakaba/wakaba/internal/application"
	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
)

type AdminHandler struct {
	admin         *application.AdminService
	messaging     *application.MessagingService
	observability *application.ObservabilityService
	logger        interface{ Info(string, ...any) }
}

func NewAdminHandler(admin *application.AdminService, messaging *application.MessagingService) *AdminHandler {
	return &AdminHandler{admin: admin, messaging: messaging}
}

func (h *AdminHandler) SetObservabilityService(obs *application.ObservabilityService) {
	h.observability = obs
}

func (h *AdminHandler) Register(app *fiber.App) {
	app.Get("/admin", h.handleDashboard)
	app.Get("/admin/businesses", h.handleBusinessesList)
	app.Post("/admin/businesses", h.handleCreateBusiness)
	app.Get("/admin/wabas", h.handleWABAsList)
	app.Post("/admin/wabas", h.handleCreateWABA)
	app.Get("/admin/phones", h.handlePhonesList)
	app.Post("/admin/phones", h.handleProvisionPhone)
	app.Get("/admin/applications", h.handleApplicationsList)
	app.Post("/admin/applications", h.handleCreateApplication)
	app.Get("/admin/credentials", h.handleCredentialsList)
	app.Post("/admin/credentials", h.handleIssueCredential)
	app.Post("/admin/credentials/:id/revoke", h.handleRevokeCredential)
	app.Get("/admin/webhooks", h.handleWebhooksList)
	app.Post("/admin/webhooks", h.handleConfigureWebhook)
	app.Get("/admin/identities", h.handleIdentitiesList)
	app.Get("/admin/conversations", h.handleConversationsList)
	app.Get("/admin/messages", h.handleMessagesList)

	// Observability reports
	app.Get("/admin/reports", h.handleReportsIndex)
	app.Get("/admin/reports/overview", h.handleSystemOverview)
	app.Get("/admin/reports/events", h.handleEventLog)
	app.Get("/admin/reports/metrics", h.handleMetrics)
}

func (h *AdminHandler) handleDashboard(c *fiber.Ctx) error {
	businesses, _ := h.admin.ListBusinesses(c.Context())
	wabas, _ := h.admin.ListWABAs(c.Context())
	phones, _ := h.admin.ListPhoneNumbers(c.Context())
	apps, _ := h.admin.ListApplications(c.Context())

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<h1 class="text-2xl font-bold text-gray-800 mb-6">Wakaba Admin</h1>
<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
  <a href="/admin/businesses" class="bg-white rounded-lg shadow p-4 hover:shadow-md transition">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Businesses</div>
  </a>
  <a href="/admin/wabas" class="bg-white rounded-lg shadow p-4 hover:shadow-md transition">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">WABAs</div>
  </a>
  <a href="/admin/phones" class="bg-white rounded-lg shadow p-4 hover:shadow-md transition">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Phone Numbers</div>
  </a>
  <a href="/admin/applications" class="bg-white rounded-lg shadow p-4 hover:shadow-md transition">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Applications</div>
  </a>
</div>
<nav class="space-y-2">
  <a href="/admin/businesses" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Businesses &amp; WABAs</a>
  <a href="/admin/phones" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Phone Numbers</a>
  <a href="/admin/applications" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Applications &amp; Credentials</a>
  <a href="/admin/webhooks" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Webhook Configuration</a>
  <a href="/admin/identities" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Consumer Identities</a>
  <a href="/admin/conversations" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Conversations</a>
  <a href="/admin/messages" class="block bg-white rounded-lg shadow p-4 hover:shadow-md">Messages</a>
  <a href="/admin/reports" class="block bg-white rounded-lg shadow p-4 hover:shadow-md border-l-4 border-[#008069]">Reports &amp; Observability</a>
</nav>
</div></body></html>`,
		len(businesses), len(wabas), len(phones), len(apps)))
}

func (h *AdminHandler) handleBusinessesList(c *fiber.Ctx) error {
	businesses, _ := h.admin.ListBusinesses(c.Context())
	wabasByBiz := make(map[string][]*domain.WABA)
	for _, b := range businesses {
		wabas, _ := h.admin.ListWABAsByBusiness(c.Context(), b.ID)
		wabasByBiz[string(b.ID)] = wabas
	}

	rows := ""
	for _, b := range businesses {
		wabaRows := ""
		for _, w := range wabasByBiz[string(b.ID)] {
			wabaRows += fmt.Sprintf(`<tr class="bg-gray-50"><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-xs text-gray-500">%s</td></tr>`, w.ID, w.Name, w.Timezone)
		}
		rows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-3 font-mono text-xs">%s</td><td class="px-4 py-3 font-medium">%s</td><td class="px-4 py-3">%s</td><td class="px-4 py-3 text-sm text-green-600">%s</td></tr>%s`, b.ID, b.Name, b.Country, b.Status, wabaRows)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Businesses - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Businesses</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/businesses" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">Name</label><input name="name" required class="border rounded px-3 py-2 text-sm"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Display Name</label><input name="display_name" class="border rounded px-3 py-2 text-sm"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Country</label><input name="country" class="border rounded px-3 py-2 text-sm" placeholder="US"></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Create Business</button>
</form>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">ID</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Name</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Country</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
</tr></thead><tbody>%s</tbody></table>
</div>
</div></body></html>`, rows))
}

func (h *AdminHandler) handleCreateBusiness(c *fiber.Ctx) error {
	name := c.FormValue("name")
	displayName := c.FormValue("display_name")
	country := c.FormValue("country")
	if name == "" {
		return c.Redirect("/admin/businesses")
	}
	_, err := h.admin.CreateBusiness(c.Context(), name, displayName, country)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/businesses")
}

func (h *AdminHandler) handleWABAsList(c *fiber.Ctx) error {
	businesses, _ := h.admin.ListBusinesses(c.Context())
	wabas, _ := h.admin.ListWABAs(c.Context())

	bizOptions := ""
	for _, b := range businesses {
		bizOptions += fmt.Sprintf(`<option value="%s">%s</option>`, b.ID, b.Name)
	}

	rows := ""
	for _, w := range wabas {
		bizName := ""
		b, _ := h.admin.GetBusiness(c.Context(), w.BusinessID)
		if b != nil {
			bizName = b.Name
		}
		rows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-xs text-gray-500">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-sm text-green-600">%s</td></tr>`, w.ID, w.Name, bizName, w.Currency, w.Status)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>WABAs - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">WhatsApp Business Accounts</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/wabas" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">Business</label><select name="business_id" class="border rounded px-3 py-2 text-sm">%s</select></div>
  <div><label class="block text-xs text-gray-500 mb-1">Name</label><input name="name" required class="border rounded px-3 py-2 text-sm"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Timezone</label><input name="timezone" class="border rounded px-3 py-2 text-sm" placeholder="America/Lima"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Currency</label><input name="currency" class="border rounded px-3 py-2 text-sm" placeholder="PEN" value="USD"></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Create WABA</button>
</form>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">ID</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Name</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Business</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Currency</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
</tr></thead><tbody>%s</tbody></table>
</div>
</div></body></html>`, bizOptions, rows))
}

func (h *AdminHandler) handleCreateWABA(c *fiber.Ctx) error {
	bizID := domain.BusinessID(c.FormValue("business_id"))
	name := c.FormValue("name")
	tz := c.FormValue("timezone")
	currency := c.FormValue("currency")
	if name == "" || bizID == "" {
		return c.Redirect("/admin/wabas")
	}
	_, err := h.admin.CreateWABA(c.Context(), bizID, name, tz, currency)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/wabas")
}

func (h *AdminHandler) handlePhonesList(c *fiber.Ctx) error {
	wabas, _ := h.admin.ListWABAs(c.Context())
	phones, _ := h.admin.ListPhoneNumbers(c.Context())

	wabaOptions := ""
	for _, w := range wabas {
		b, _ := h.admin.GetBusiness(c.Context(), w.BusinessID)
		label := w.Name
		if b != nil {
			label = b.Name + " / " + w.Name
		}
		wabaOptions += fmt.Sprintf(`<option value="%s">%s</option>`, w.ID, label)
	}

	rows := ""
	for _, p := range phones {
		wabaName := ""
		w, _ := h.admin.GetWABA(c.Context(), p.WABAID)
		if w != nil {
			wabaName = w.Name
		}
		rows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2 font-mono">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-xs text-gray-500">%s</td><td class="px-4 py-2 text-sm text-green-600">%s</td></tr>`, p.ID, p.Phone, p.DisplayName, wabaName, p.Status)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Phone Numbers - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Business Phone Numbers</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/phones" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">WABA</label><select name="waba_id" class="border rounded px-3 py-2 text-sm">%s</select></div>
  <div><label class="block text-xs text-gray-500 mb-1">Phone Number</label><input name="phone" required class="border rounded px-3 py-2 text-sm font-mono" placeholder="+5115550100"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Display Name</label><input name="display_name" class="border rounded px-3 py-2 text-sm"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Country</label><input name="country" class="border rounded px-3 py-2 text-sm" placeholder="PE"></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Provision Number</button>
</form>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">ID</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Phone</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Display Name</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">WABA</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
</tr></thead><tbody>%s</tbody></table>
</div>
</div></body></html>`, wabaOptions, rows))
}

func (h *AdminHandler) handleProvisionPhone(c *fiber.Ctx) error {
	wabaID := domain.WABAID(c.FormValue("waba_id"))
	phone := domain.NormalizePhoneNumber(c.FormValue("phone"))
	displayName := c.FormValue("display_name")
	country := c.FormValue("country")
	if wabaID == "" {
		return c.Redirect("/admin/phones")
	}
	_, err := h.admin.ProvisionPhoneNumber(c.Context(), wabaID, phone, displayName, country)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/phones")
}

func (h *AdminHandler) handleApplicationsList(c *fiber.Ctx) error {
	businesses, _ := h.admin.ListBusinesses(c.Context())
	apps, _ := h.admin.ListApplications(c.Context())

	bizOptions := ""
	for _, b := range businesses {
		bizOptions += fmt.Sprintf(`<option value="%s">%s</option>`, b.ID, b.Name)
	}

	rows := ""
	for _, a := range apps {
		bizName := ""
		b, _ := h.admin.GetBusiness(c.Context(), a.BusinessID)
		if b != nil {
			bizName = b.Name
		}
		rows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-xs text-gray-500">%s</td><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2 text-sm text-green-600">%s</td></tr>`, a.ID, a.Name, bizName, a.AppSecret[:16]+"...", a.Status)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Applications - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Applications</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/applications" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">Business</label><select name="business_id" class="border rounded px-3 py-2 text-sm">%s</select></div>
  <div><label class="block text-xs text-gray-500 mb-1">Name</label><input name="name" required class="border rounded px-3 py-2 text-sm"></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Create Application</button>
</form>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">ID</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Name</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Business</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">App Secret</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
</tr></thead><tbody>%s</tbody></table>
</div>
<div class="mt-6"><a href="/admin/credentials" class="text-[#008069] hover:underline">Manage Credentials &rarr;</a></div>
</div></body></html>`, bizOptions, rows))
}

func (h *AdminHandler) handleCreateApplication(c *fiber.Ctx) error {
	bizID := domain.BusinessID(c.FormValue("business_id"))
	name := c.FormValue("name")
	if name == "" || bizID == "" {
		return c.Redirect("/admin/applications")
	}
	_, err := h.admin.CreateApplication(c.Context(), bizID, name)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/applications")
}

func (h *AdminHandler) handleCredentialsList(c *fiber.Ctx) error {
	apps, _ := h.admin.ListApplications(c.Context())

	appOptions := ""
	for _, a := range apps {
		appOptions += fmt.Sprintf(`<option value="%s">%s</option>`, a.ID, a.Name)
	}

	type credRow struct {
		CredID     string
		AppID      string
		AppName    string
		TokenHash  string
		Status     string
	}
	var rows []credRow
	for _, a := range apps {
		creds, _ := h.admin.ListCredentials(c.Context(), a.ID)
		for _, cr := range creds {
			hashPreview := cr.TokenHash
			if len(hashPreview) > 16 {
				hashPreview = hashPreview[:16] + "..."
			}
			rows = append(rows, credRow{
				CredID:    string(cr.ID),
				AppID:     string(cr.ApplicationID),
				AppName:   a.Name,
				TokenHash: hashPreview,
				Status:    cr.Status,
			})
		}
	}

	rowHTML := ""
	for _, r := range rows {
		revokeBtn := ""
		if r.Status == "active" {
			revokeBtn = fmt.Sprintf(`<form method="POST" action="/admin/credentials/%s/revoke" style="display:inline"><button class="text-red-500 text-xs hover:underline" onclick="return confirm('Revoke this token?')">Revoke</button></form>`, r.CredID)
		}
		rowHTML += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2">%s</td><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2 text-sm">%s</td><td class="px-4 py-2">%s</td></tr>`, r.CredID, r.AppName, r.TokenHash, r.Status, revokeBtn)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Credentials - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">API Credentials</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/credentials" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">Application</label><select name="application_id" class="border rounded px-3 py-2 text-sm">%s</select></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Issue Token</button>
</form>
<div id="new-token" class="hidden bg-green-50 border border-green-200 rounded m-4 p-4">
  <div class="text-sm font-medium text-green-800 mb-1">New token issued!</div>
  <div class="text-xs text-green-700">Copy this token now. It will not be shown again.</div>
  <code id="token-value" class="block mt-2 p-2 bg-white rounded text-sm font-mono break-all border"></code>
</div>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">ID</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Application</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Token Hash</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Action</th>
</tr></thead><tbody>%s</tbody></table>
</div>
</div></body></html>`, appOptions, rowHTML))
}

func (h *AdminHandler) handleIssueCredential(c *fiber.Ctx) error {
	appID := domain.ApplicationID(c.FormValue("application_id"))
	if appID == "" {
		return c.Redirect("/admin/credentials")
	}
	_, token, err := h.admin.IssueCredential(c.Context(), appID)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Token Issued - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen flex items-center justify-center">
<div class="bg-white rounded-lg shadow-lg p-8 max-w-lg w-full">
  <h2 class="text-xl font-bold text-gray-800 mb-4">Token Issued</h2>
  <div class="bg-yellow-50 border border-yellow-200 rounded p-4 mb-4">
    <div class="text-sm font-medium text-yellow-800">Copy this token now. It will not be shown again.</div>
  </div>
  <div class="bg-gray-100 rounded p-3 font-mono text-sm break-all mb-6">%s</div>
  <div class="flex gap-3">
    <a href="/admin/credentials" class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Back to Credentials</a>
    <a href="/admin" class="bg-gray-200 text-gray-700 px-4 py-2 rounded text-sm font-medium hover:bg-gray-300">Dashboard</a>
  </div>
</div></body></html>`, token))
}

func (h *AdminHandler) handleRevokeCredential(c *fiber.Ctx) error {
	credID := domain.CredentialID(c.Params("id"))
	if err := h.admin.RevokeCredential(c.Context(), credID); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/credentials")
}

func (h *AdminHandler) handleWebhooksList(c *fiber.Ctx) error {
	apps, _ := h.admin.ListApplications(c.Context())

	appOptions := ""
	for _, a := range apps {
		appOptions += fmt.Sprintf(`<option value="%s">%s</option>`, a.ID, a.Name)
	}

	type whRow struct {
		AppID      string
		AppName    string
		CallbackURL string
		VerifyToken string
		Enabled    bool
	}
	var rows []whRow
	for _, a := range apps {
		wh, _ := h.admin.GetWebhookConfig(c.Context(), a.ID)
		if wh != nil {
			rows = append(rows, whRow{
				AppID:       string(a.ID),
				AppName:     a.Name,
				CallbackURL: wh.CallbackURL,
				VerifyToken: wh.VerifyToken,
				Enabled:     wh.Enabled,
			})
		}
	}

	rowHTML := ""
	for _, r := range rows {
		status := `<span class="text-green-600">Enabled</span>`
		if !r.Enabled {
			status = `<span class="text-red-500">Disabled</span>`
		}
		rowHTML += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2">%s</td><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2 font-mono text-xs">%s</td><td class="px-4 py-2">%s</td></tr>`, r.AppName, r.CallbackURL, r.VerifyToken, status)
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Webhooks - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Webhook Configuration</h1>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden mb-6">
<form method="POST" action="/admin/webhooks" class="p-4 flex gap-3 items-end border-b">
  <div><label class="block text-xs text-gray-500 mb-1">Application</label><select name="application_id" class="border rounded px-3 py-2 text-sm">%s</select></div>
  <div><label class="block text-xs text-gray-500 mb-1">Callback URL</label><input name="callback_url" required class="border rounded px-3 py-2 text-sm font-mono" placeholder="http://localhost:3000/webhooks/whatsapp"></div>
  <div><label class="block text-xs text-gray-500 mb-1">Verify Token</label><input name="verify_token" required class="border rounded px-3 py-2 text-sm" placeholder="my-verify-token"></div>
  <button class="bg-[#008069] text-white px-4 py-2 rounded text-sm font-medium hover:bg-[#006d5b]">Save Webhook</button>
</form>
<table class="w-full text-sm"><thead class="bg-gray-100"><tr>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Application</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Callback URL</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Verify Token</th>
  <th class="px-4 py-2 text-left text-xs text-gray-500">Status</th>
</tr></thead><tbody>%s</tbody></table>
</div>
</div></body></html>`, appOptions, rowHTML))
}

func (h *AdminHandler) handleConfigureWebhook(c *fiber.Ctx) error {
	appID := domain.ApplicationID(c.FormValue("application_id"))
	callbackURL := c.FormValue("callback_url")
	verifyToken := c.FormValue("verify_token")
	if appID == "" || callbackURL == "" {
		return c.Redirect("/admin/webhooks")
	}
	_, err := h.admin.ConfigureWebhook(c.Context(), appID, callbackURL, verifyToken)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	return c.Redirect("/admin/webhooks")
}

func (h *AdminHandler) handleIdentitiesList(c *fiber.Ctx) error {
	_ = h.messaging
	return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Identities - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Consumer Identities</h1>
</div>
<p class="text-gray-500 text-sm">Consumer identities are created through the <a href="/" class="text-[#008069] hover:underline">Virtual Client</a>.</p>
</div></body></html>`)
}

func (h *AdminHandler) handleConversationsList(c *fiber.Ctx) error {
	return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Conversations - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Conversations</h1>
</div>
<p class="text-gray-500 text-sm">Conversations are managed through the Virtual Client and API.</p>
</div></body></html>`)
}

func (h *AdminHandler) handleMessagesList(c *fiber.Ctx) error {
	return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Messages - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Messages</h1>
</div>
<p class="text-gray-500 text-sm">Message inspection available through the API and Virtual Client.</p>
</div></body></html>`)
}

func (h *AdminHandler) handleReportsIndex(c *fiber.Ctx) error {
	return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Reports - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Reports</h1>
</div>
<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
  <a href="/admin/reports/overview" class="bg-white rounded-lg shadow p-6 hover:shadow-md transition">
    <div class="text-lg font-semibold text-gray-800">System Overview</div>
    <div class="text-sm text-gray-500 mt-1">World summary, event counts, failure rate</div>
  </a>
  <a href="/admin/reports/events" class="bg-white rounded-lg shadow p-6 hover:shadow-md transition">
    <div class="text-lg font-semibold text-gray-800">Event Log</div>
    <div class="text-sm text-gray-500 mt-1">Recent events with filtering</div>
  </a>
  <a href="/admin/reports/metrics" class="bg-white rounded-lg shadow p-6 hover:shadow-md transition">
    <div class="text-lg font-semibold text-gray-800">Metrics</div>
    <div class="text-sm text-gray-500 mt-1">Event distribution and outcomes</div>
  </a>
</div>
</div></body></html>`)
}

func (h *AdminHandler) handleSystemOverview(c *fiber.Ctx) error {
	if h.observability == nil {
		return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>System Overview - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<h1 class="text-2xl font-bold text-gray-800 mb-4">System Overview</h1>
<p class="text-gray-500">Observability service not configured.</p>
</div></body></html>`)
	}

	summary, err := h.observability.GetWorldSummary(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Type("html").SendString(func() string {
		base := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>System Overview - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin/reports" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">System Overview</h1>
</div>
<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Consumers</div>
  </div>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Events (24h)</div>
  </div>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-red-600">%d</div>
    <div class="text-sm text-gray-500">Failures (24h)</div>
  </div>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-[#008069]">%.1f%%</div>
    <div class="text-sm text-gray-500">Failure Rate</div>
  </div>
</div>
<h2 class="text-lg font-semibold text-gray-800 mb-3">Events by Type (24h)</h2>
<div class="bg-white rounded-lg shadow overflow-hidden mb-8">
<table class="w-full text-sm">
<thead class="bg-gray-50"><tr><th class="px-4 py-2 text-left">Event Type</th><th class="px-4 py-2 text-right">Count</th></tr></thead>
<tbody>`, summary.ConsumerCount, summary.EventCount, summary.FailureCount,
			func() float64 {
				if summary.EventCount > 0 {
					return float64(summary.FailureCount) / float64(summary.EventCount) * 100
				}
				return 0
			}())

		typeRows := ""
		for et, count := range summary.EventsByType {
			typeRows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-right font-mono">%d</td></tr>`, et, count)
		}
		if typeRows == "" {
			typeRows = `<tr><td colspan="2" class="px-4 py-2 text-gray-400">No events recorded</td></tr>`
		}

		outcomeRows := ""
		for outcome, count := range summary.EventsByOutcome {
			outcomeRows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-right font-mono">%d</td></tr>`, outcome, count)
		}
		if outcomeRows == "" {
			outcomeRows = `<tr><td colspan="2" class="px-4 py-2 text-gray-400">No outcomes recorded</td></tr>`
		}

		return base + typeRows + `</tbody></table></div>
<h2 class="text-lg font-semibold text-gray-800 mb-3">Outcomes (24h)</h2>
<div class="bg-white rounded-lg shadow overflow-hidden">
<table class="w-full text-sm">
<thead class="bg-gray-50"><tr><th class="px-4 py-2 text-left">Outcome</th><th class="px-4 py-2 text-right">Count</th></tr></thead>
<tbody>` + outcomeRows + `</tbody></table></div>
</div></body></html>`
	}())
}

func (h *AdminHandler) handleEventLog(c *fiber.Ctx) error {
	if h.observability == nil {
		return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Event Log - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<h1 class="text-2xl font-bold text-gray-800 mb-4">Event Log</h1>
<p class="text-gray-500">Observability service not configured.</p>
</div></body></html>`)
	}

	result, err := h.observability.ListEvents(c.Context(), ports.EventFilter{Limit: 50})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Type("html").SendString(func() string {
		base := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Event Log - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin/reports" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Event Log</h1>
  <span class="text-sm text-gray-500">%d total</span>
</div>
<div class="bg-white rounded-lg shadow overflow-hidden">
<table class="w-full text-sm">
<thead class="bg-gray-50">
<tr>
  <th class="px-4 py-2 text-left">Time</th>
  <th class="px-4 py-2 text-left">Type</th>
  <th class="px-4 py-2 text-left">Source</th>
  <th class="px-4 py-2 text-left">Outcome</th>
  <th class="px-4 py-2 text-left">Trace ID</th>
  <th class="px-4 py-2 text-left">Message ID</th>
</tr>
</thead>
<tbody>`, result.Total)

		rows := ""
		for _, ev := range result.Events {
			traceShort := string(ev.TraceID)
			if len(traceShort) > 12 {
				traceShort = traceShort[:12] + "..."
			}
			msgShort := ev.MessageID
			if len(msgShort) > 12 {
				msgShort = msgShort[:12] + "..."
			}
			outcomeClass := "text-gray-600"
			if ev.Outcome == domain.EventOutcomeFailure {
				outcomeClass = "text-red-600 font-semibold"
			}
			rows += fmt.Sprintf(`<tr class="border-t hover:bg-gray-50">
<td class="px-4 py-2 font-mono text-xs">%s</td>
<td class="px-4 py-2">%s</td>
<td class="px-4 py-2 text-gray-500">%s</td>
<td class="px-4 py-2 %s">%s</td>
<td class="px-4 py-2 font-mono text-xs text-gray-500">%s</td>
<td class="px-4 py-2 font-mono text-xs text-gray-500">%s</td>
</tr>`, ev.Timestamp.Format("15:04:05.000"), ev.Type, ev.Source, outcomeClass, ev.Outcome, traceShort, msgShort)
		}
		if rows == "" {
			rows = `<tr><td colspan="6" class="px-4 py-8 text-center text-gray-400">No events recorded</td></tr>`
		}

		return base + rows + `</tbody></table></div>
</div></body></html>`
	}())
}

func (h *AdminHandler) handleMetrics(c *fiber.Ctx) error {
	if h.observability == nil {
		return c.Type("html").SendString(`<!DOCTYPE html>
<html><head><title>Metrics - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<h1 class="text-2xl font-bold text-gray-800 mb-4">Metrics</h1>
<p class="text-gray-500">Observability service not configured.</p>
</div></body></html>`)
	}

	summary, err := h.observability.GetMetricsSummary(c.Context(), 24*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Type("html").SendString(func() string {
		base := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Metrics - Wakaba Admin</title>
<script src="https://cdn.tailwindcss.com"></script>
</head><body class="bg-gray-50 min-h-screen">
<div class="max-w-6xl mx-auto px-6 py-8">
<div class="flex items-center gap-4 mb-6">
  <a href="/admin/reports" class="text-gray-500 hover:text-gray-700">&larr; Back</a>
  <h1 class="text-2xl font-bold text-gray-800">Metrics</h1>
  <span class="text-sm text-gray-500">Last %s</span>
</div>
<div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8">
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Total Events</div>
  </div>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-red-600">%.1f%%</div>
    <div class="text-sm text-gray-500">Failure Rate</div>
  </div>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="text-2xl font-bold text-[#008069]">%d</div>
    <div class="text-sm text-gray-500">Event Types</div>
  </div>
</div>
<h2 class="text-lg font-semibold text-gray-800 mb-3">Event Distribution</h2>
<div class="bg-white rounded-lg shadow overflow-hidden mb-8">
<table class="w-full text-sm">
<thead class="bg-gray-50"><tr><th class="px-4 py-2 text-left">Event Type</th><th class="px-4 py-2 text-right">Count</th><th class="px-4 py-2 text-right">%%</th></tr></thead>
<tbody>`, summary.Period, summary.TotalEvents, summary.FailureRate, len(summary.EventsByType))

		typeRows := ""
		for et, count := range summary.EventsByType {
			pct := float64(0)
			if summary.TotalEvents > 0 {
				pct = float64(count) / float64(summary.TotalEvents) * 100
			}
			typeRows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-right font-mono">%d</td><td class="px-4 py-2 text-right text-gray-500">%.1f</td></tr>`, et, count, pct)
		}
		if typeRows == "" {
			typeRows = `<tr><td colspan="3" class="px-4 py-2 text-gray-400">No data</td></tr>`
		}

		outcomeRows := ""
		for outcome, count := range summary.EventsByOutcome {
			pct := float64(0)
			if summary.TotalEvents > 0 {
				pct = float64(count) / float64(summary.TotalEvents) * 100
			}
			outcomeRows += fmt.Sprintf(`<tr class="border-t"><td class="px-4 py-2">%s</td><td class="px-4 py-2 text-right font-mono">%d</td><td class="px-4 py-2 text-right text-gray-500">%.1f</td></tr>`, outcome, count, pct)
		}
		if outcomeRows == "" {
			outcomeRows = `<tr><td colspan="3" class="px-4 py-2 text-gray-400">No data</td></tr>`
		}

		return base + typeRows + `</tbody></table></div>
<h2 class="text-lg font-semibold text-gray-800 mb-3">Outcomes</h2>
<div class="bg-white rounded-lg shadow overflow-hidden">
<table class="w-full text-sm">
<thead class="bg-gray-50"><tr><th class="px-4 py-2 text-left">Outcome</th><th class="px-4 py-2 text-right">Count</th><th class="px-4 py-2 text-right">%%</th></tr></thead>
<tbody>` + outcomeRows + `</tbody></table></div>
</div></body></html>`
	}())
}
