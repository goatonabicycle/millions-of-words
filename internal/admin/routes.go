package admin

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, h *Handler) {
	admin := e.Group("/admin")
	admin.GET("", h.AdminHandler)
	admin.POST("/auth", h.AdminAuthHandler)
	admin.GET("/content/import", h.AdminImportHandler)
	admin.POST("/import", h.ImportAlbumHandler)
	admin.POST("/fetch/metal-archives", h.FetchMetalArchivesHandler)
	admin.POST("/validate/metal-archives-url", h.ValidateMetalArchivesUrlHandler)
	admin.GET("/logout", h.LogoutHandler)
	admin.GET("/content/albums", h.AlbumListHandler)
}
