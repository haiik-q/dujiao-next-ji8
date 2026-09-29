package xcheckhttp

import (
	"context"
	"errors"

	"github.com/dujiao-next/internal/modules/xcheck/domain"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// Checker 执行 X ID 自检。
type Checker interface {
	Check(ctx context.Context, raw string) (domain.Result, error)
}

// Handler 处理公开的 X ID 自检请求。
type Handler struct {
	checker Checker
}

// NewHandler 创建处理器。
func NewHandler(checker Checker) *Handler {
	return &Handler{checker: checker}
}

type checkRequest struct {
	Handle string `json:"handle" binding:"required,max=200"`
}

// Check POST /api/v1/public/x-check
func (h *Handler) Check(c *gin.Context) {
	var req checkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}
	result, err := h.checker.Check(c.Request.Context(), req.Handle)
	switch {
	case errors.Is(err, domain.ErrInvalidHandle):
		ginutil.RespondError(c, response.CodeBadRequest, "error.x_check_invalid_handle", nil)
	case err != nil:
		ginutil.RespondError(c, response.CodeInternal, "error.x_check_unavailable", err)
	default:
		response.Success(c, result)
	}
}
