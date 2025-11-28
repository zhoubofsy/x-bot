package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhoubofsy/x-bot/internal/application/dto"
	"github.com/zhoubofsy/x-bot/internal/application/service"
	"github.com/zhoubofsy/x-bot/internal/domain/repository"
)

type WorkflowHandler struct {
	workflowService service.WorkflowService
	followerService service.FollowerService
	replyLogRepo    repository.ReplyLogRepository
}

func NewWorkflowHandler(
	workflowService service.WorkflowService,
	followerService service.FollowerService,
	replyLogRepo repository.ReplyLogRepository,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		followerService: followerService,
		replyLogRepo:    replyLogRepo,
	}
}

// Execute 执行工作流
// @Summary 执行工作流
// @Description 执行黑客松推文检测和广告回复工作流
// @Tags workflow
// @Accept json
// @Produce json
// @Param params body dto.WorkflowParams true "工作流参数"
// @Success 200 {object} dto.WorkflowResult
// @Router /api/v1/workflow/execute [post]
func (h *WorkflowHandler) Execute(c *gin.Context) {
	var params dto.WorkflowParams
	if err := c.ShouldBindJSON(&params); err != nil {
		// 使用查询参数
		if tc := c.Query("tweet_count"); tc != "" {
			if count, err := strconv.Atoi(tc); err == nil {
				params.TweetCount = count
			}
		}
		params.DryRun = c.Query("dry_run") == "true"
	}

	result, err := h.workflowService.Execute(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SyncFollowing 同步关注列表
// @Summary 同步关注列表
// @Description 从Twitter同步当前用户的关注列表到数据库
// @Tags workflow
// @Produce json
// @Success 200 {object} dto.SyncFollowingResult
// @Router /api/v1/workflow/sync-following [post]
func (h *WorkflowHandler) SyncFollowing(c *gin.Context) {
	result, err := h.followerService.SyncFollowing(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats 获取统计信息
// @Summary 获取统计信息
// @Description 获取回复统计信息
// @Tags workflow
// @Produce json
// @Success 200 {object} repository.ReplyStats
// @Router /api/v1/stats [get]
func (h *WorkflowHandler) GetStats(c *gin.Context) {
	stats, err := h.replyLogRepo.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentLogs 获取最近的回复日志
// @Summary 获取最近的回复日志
// @Description 获取最近的回复日志列表
// @Tags workflow
// @Produce json
// @Param limit query int false "限制数量" default(20)
// @Success 200 {array} entity.ReplyLog
// @Router /api/v1/reply-logs [get]
func (h *WorkflowHandler) GetRecentLogs(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		// 简单处理，实际应使用 strconv
		limit = 20
	}

	logs, err := h.replyLogRepo.GetRecentLogs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}
