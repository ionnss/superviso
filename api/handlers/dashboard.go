package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func DashboardHandler(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{
		"ShowRoleWarning": true,
		// ... outros dados do dashboard
	})
}
