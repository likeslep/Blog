// controller/tag_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type TagController struct {
	tagService *service.TagService
}

func NewTagController(tagService *service.TagService) *TagController {
	return &TagController{tagService: tagService}
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// Create 创建标签（需要管理员权限）
func (ctrl *TagController) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	tag, err := ctrl.tagService.Create(req.Name, "")
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "标签创建成功", tag)
}

// Update 更新标签（需要管理员权限）
func (ctrl *TagController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的标签ID")
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	tag, err := ctrl.tagService.Update(uint(id), req.Name)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "标签更新成功", tag)
}

// Delete 删除标签（需要管理员权限）
func (ctrl *TagController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的标签ID")
		return
	}

	err = ctrl.tagService.Delete(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "标签删除成功", nil)
}

// GetByID 获取标签详情（公开）
func (ctrl *TagController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的标签ID")
		return
	}

	tag, err := ctrl.tagService.GetByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tag)
}

// GetBySlug 通过Slug获取标签详情（公开）
func (ctrl *TagController) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ValidationError(c, "无效的标签Slug")
		return
	}

	tag, err := ctrl.tagService.GetBySlug(slug)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tag)
}

// List 获取标签列表（公开）
func (ctrl *TagController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	tags, total, err := ctrl.tagService.List(page, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        tags,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAll 获取所有标签（不分页，用于下拉选择）
func (ctrl *TagController) GetAll(c *gin.Context) {
	tags, err := ctrl.tagService.GetAll()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tags)
}

// GetPopular 获取热门标签
func (ctrl *TagController) GetPopular(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tags, err := ctrl.tagService.GetPopular(limit)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tags)
}
