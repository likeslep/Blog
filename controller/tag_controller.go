// controller/tag_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

// SwaggerTagResponse 标签响应（用于 Swagger 文档）
type SwaggerTagResponse struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"Go"`
	Slug      string `json:"slug" example:"go"`
	PostCount int    `json:"post_count" example:"5"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

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

// Create 创建标签
// @Summary      创建标签
// @Description  创建新标签（需要管理员权限）
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateTagRequest true "标签信息"
// @Success      200  {object}  utils.Response{data=SwaggerTagResponse} "创建成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      409  {object}  utils.Response "标签已存在"
// @Router       /admin/tags [post]
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

// Update 更新标签
// @Summary      更新标签
// @Description  更新标签信息（需要管理员权限）
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "标签ID"
// @Param        request body UpdateTagRequest true "标签信息"
// @Success      200  {object}  utils.Response{data=SwaggerTagResponse} "更新成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "标签不存在"
// @Failure      409  {object}  utils.Response "标签已存在"
// @Router       /admin/tags/{id} [put]
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

// Delete 删除标签
// @Summary      删除标签
// @Description  删除标签（需要管理员权限）
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "标签ID"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "标签不存在"
// @Router       /admin/tags/{id} [delete]
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

// GetByID 获取标签详情
// @Summary      获取标签详情
// @Description  根据ID获取标签详情
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id path int true "标签ID"
// @Success      200  {object}  utils.Response{data=SwaggerTagResponse} "获取成功"
// @Failure      404  {object}  utils.Response "标签不存在"
// @Router       /tags/{id} [get]
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

// GetBySlug 通过Slug获取标签详情
// @Summary      通过Slug获取标签详情
// @Description  根据Slug获取标签详情
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        slug path string true "标签Slug"
// @Success      200  {object}  utils.Response{data=SwaggerTagResponse} "获取成功"
// @Failure      404  {object}  utils.Response "标签不存在"
// @Router       /tags/slug/{slug} [get]
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

// List 获取标签列表
// @Summary      获取标签列表
// @Description  分页获取标签列表
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=SwaggerTagListData} "获取成功"
// @Router       /tags [get]
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

// GetAll 获取所有标签
// @Summary      获取所有标签
// @Description  获取所有标签（不分页，用于下拉选择）
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]SwaggerTagResponse} "获取成功"
// @Router       /tags/all [get]
func (ctrl *TagController) GetAll(c *gin.Context) {
	tags, err := ctrl.tagService.GetAll()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tags)
}

// GetPopular 获取热门标签
// @Summary      获取热门标签
// @Description  获取文章数量最多的标签
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        limit query int false "返回数量" default(10) maximum(50)
// @Success      200  {object}  utils.Response{data=[]SwaggerTagResponse} "获取成功"
// @Router       /tags/popular [get]
func (ctrl *TagController) GetPopular(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	tags, err := ctrl.tagService.GetPopular(limit)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, tags)
}
