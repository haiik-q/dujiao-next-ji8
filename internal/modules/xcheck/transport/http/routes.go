package xcheckhttp

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册公开的 X ID 自检接口（调用方负责挂载按 IP 的限流）。
func RegisterPublicRoutes(public gin.IRoutes, handler *Handler) {
	public.POST("/x-check", handler.Check)
}
