package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// AddUserRequest 添加用户请求
type AddUserRequest struct {
	TwitterUserID string `json:"twitter_user_id" binding:"required"`
	Username      string `json:"username" binding:"required"`
	DisplayName   string `json:"display_name"`
}

// List 获取所有监控用户
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.userRepo.GetAllActiveUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Add 手动添加监控用户
func (h *UserHandler) Add(c *gin.Context) {
	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否已存在
	existing, _ := h.userRepo.GetByTwitterID(c.Request.Context(), req.TwitterUserID)
	if existing != nil {
		// 如果已存在但被禁用，重新激活
		if !existing.IsActive {
			h.userRepo.UpdateActiveStatus(c.Request.Context(), req.TwitterUserID, true)
			existing.IsActive = true
			c.JSON(http.StatusOK, gin.H{
				"message": "用户已重新激活",
				"user":    existing,
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "用户已存在"})
		return
	}

	user := &entity.FollowedUser{
		TwitterUserID: req.TwitterUserID,
		Username:      req.Username,
		DisplayName:   req.DisplayName,
		IsActive:      true,
	}

	if err := h.userRepo.Save(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户添加成功",
		"user":    user,
	})
}

// BatchAdd 批量添加监控用户
func (h *UserHandler) BatchAdd(c *gin.Context) {
	var users []AddUserRequest
	if err := c.ShouldBindJSON(&users); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var added, skipped int
	for _, req := range users {
		existing, _ := h.userRepo.GetByTwitterID(c.Request.Context(), req.TwitterUserID)
		if existing != nil {
			skipped++
			continue
		}

		user := &entity.FollowedUser{
			TwitterUserID: req.TwitterUserID,
			Username:      req.Username,
			DisplayName:   req.DisplayName,
			IsActive:      true,
		}
		if err := h.userRepo.Save(c.Request.Context(), user); err == nil {
			added++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量添加完成",
		"added":   added,
		"skipped": skipped,
	})
}

// Delete 删除监控用户
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

// UpdateStatus 更新用户状态
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	twitterID := c.Param("twitter_id")
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userRepo.UpdateActiveStatus(c.Request.Context(), twitterID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

