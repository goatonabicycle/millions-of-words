package admin

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, h *Handler) {
	admin := e.Group("/admin")
	admin.GET("", h.AdminHandler)
	admin.POST("/auth", h.AdminAuthHandler)
	admin.GET("/content/editor", h.AdminEditorHandler)
	admin.GET("/content/import", h.AdminImportHandler)
	admin.POST("/update/album", h.UpdateAlbumHandler)
	admin.POST("/update/track", h.UpdateTrackHandler)
	admin.POST("/import", h.ImportAlbumHandler)
	admin.GET("/content/editor/album", h.AdminAlbumEditorHandler)
	admin.POST("/fetch/metal-archives", h.FetchMetalArchivesHandler)
	admin.POST("/validate/metal-archives-url", h.ValidateMetalArchivesUrlHandler)
}
