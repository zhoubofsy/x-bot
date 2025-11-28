package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhoubofsy/x-bot/internal/domain/entity"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
)

type AdCopyHandler struct {
	adCopyRepo repository.AdCopyRepository
}

func NewAdCopyHandler(adCopyRepo repository.AdCopyRepository) *AdCopyHandler {
	return &AdCopyHandler{adCopyRepo: adCopyRepo}
}

// List 获取所有广告文案
// @Summary 获取广告文案列表
// @Tags ad-copies
// @Produce json
// @Success 200 {array} entity.AdCopy
// @Router /api/v1/ad-copies [get]
func (h *AdCopyHandler) List(c *gin.Context) {
	adCopies, err := h.adCopyRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, adCopies)
}

// Get 获取单个广告文案
// @Summary 获取单个广告文案
// @Tags ad-copies
// @Produce json
// @Param id path int true "广告文案ID"
// @Success 200 {object} entity.AdCopy
// @Router /api/v1/ad-copies/{id} [get]
func (h *AdCopyHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	adCopy, err := h.adCopyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ad copy not found"})
		return
	}
	c.JSON(http.StatusOK, adCopy)
}

// Create 创建广告文案
// @Summary 创建广告文案
// @Tags ad-copies
// @Accept json
// @Produce json
// @Param input body entity.CreateAdCopyInput true "广告文案信息"
// @Success 201 {object} entity.AdCopy
// @Router /api/v1/ad-copies [post]
func (h *AdCopyHandler) Create(c *gin.Context) {
	var input entity.CreateAdCopyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adCopy := &entity.AdCopy{
		Name:     input.Name,
		Content:  input.Content,
		Category: input.Category,
		Priority: input.Priority,
		IsActive: true,
	}

	if adCopy.Category == "" {
		adCopy.Category = "hackathon"
	}

	if err := h.adCopyRepo.Save(c.Request.Context(), adCopy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, adCopy)
}

// Update 更新广告文案
// @Summary 更新广告文案
// @Tags ad-copies
// @Accept json
// @Produce json
// @Param id path int true "广告文案ID"
// @Param input body entity.UpdateAdCopyInput true "更新信息"
// @Success 200 {object} entity.AdCopy
// @Router /api/v1/ad-copies/{id} [put]
func (h *AdCopyHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	adCopy, err := h.adCopyRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ad copy not found"})
		return
	}

	var input entity.UpdateAdCopyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != nil {
		adCopy.Name = *input.Name
	}
	if input.Content != nil {
		adCopy.Content = *input.Content
	}
	if input.Category != nil {
		adCopy.Category = *input.Category
	}
	if input.Priority != nil {
		adCopy.Priority = *input.Priority
	}
	if input.IsActive != nil {
		adCopy.IsActive = *input.IsActive
	}

	if err := h.adCopyRepo.Update(c.Request.Context(), adCopy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, adCopy)
}

// Delete 删除广告文案
// @Summary 删除广告文案
// @Tags ad-copies
// @Param id path int true "广告文案ID"
// @Success 204
// @Router /api/v1/ad-copies/{id} [delete]
func (h *AdCopyHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.adCopyRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

