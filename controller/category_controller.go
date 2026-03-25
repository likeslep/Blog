// controller/category_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SwaggerCategoryResponse 分类响应（用于 Swagger 文档）
type SwaggerCategoryResponse struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"技术"`
	Slug        string `json:"slug" example:"ji-shu"`
	Description string `json:"description" example:"技术相关文章"`
	PostCount   int    `json:"post_count" example:"10"`
	SortOrder   int    `json:"sort_order" example:"1"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// SwaggerCategoryListData 分类列表响应（用于 Swagger 文档）
type SwaggerCategoryListData struct {
	List       []SwaggerCategoryResponse `json:"list"`
	Total      int64                     `json:"total" example:"100"`
	Page       int                       `json:"page" example:"1"`
	PageSize   int                       `json:"page_size" example:"10"`
	TotalPages int64                     `json:"total_pages" example:"10"`
}

type CategoryController struct {
	categoryService *service.CategoryService
}

func NewCategoryController(categoryService *service.CategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	SortOrder   int    `json:"sort_order"`
}

// Create 创建分类
// @Summary      创建分类
// @Description  创建新分类（需要管理员权限）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateCategoryRequest true "分类信息"
// @Success      200  {object}  utils.Response{data=SwaggerCategoryResponse} "创建成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      409  {object}  utils.Response "分类已存在"
// @Router       /admin/categories [post]
func (ctrl *CategoryController) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	category, err := ctrl.categoryService.Create(req.Name, req.Description, req.SortOrder)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "分类创建成功", category)
}

// Update 更新分类
// @Summary      更新分类
// @Description  更新分类信息（需要管理员权限）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "分类ID"
// @Param        request body UpdateCategoryRequest true "分类信息"
// @Success      200  {object}  utils.Response{data=SwaggerCategoryResponse} "更新成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "分类不存在"
// @Failure      409  {object}  utils.Response "分类已存在"
// @Router       /admin/categories/{id} [put]
func (ctrl *CategoryController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的分类ID")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, "参数验证失败: "+err.Error())
		return
	}

	category, err := ctrl.categoryService.Update(uint(id), req.Name, req.Description, req.SortOrder)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "分类更新成功", category)
}

// Delete 删除分类
// @Summary      删除分类
// @Description  删除分类（需要管理员权限，只能删除没有文章的空白分类）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "分类ID"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "分类不存在"
// @Failure      409  {object}  utils.Response "分类下还有文章，无法删除"
// @Router       /admin/categories/{id} [delete]
func (ctrl *CategoryController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的分类ID")
		return
	}

	err = ctrl.categoryService.Delete(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "分类删除成功", nil)
}

// GetByID 获取分类详情
// @Summary      获取分类详情
// @Description  根据ID获取分类详情
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        id path int true "分类ID"
// @Success      200  {object}  utils.Response{data=SwaggerCategoryResponse} "获取成功"
// @Failure      404  {object}  utils.Response "分类不存在"
// @Router       /categories/{id} [get]
func (ctrl *CategoryController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的分类ID")
		return
	}

	category, err := ctrl.categoryService.GetByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, category)
}

// GetBySlug 通过Slug获取分类详情
// @Summary      通过Slug获取分类详情
// @Description  根据Slug获取分类详情
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        slug path string true "分类Slug"
// @Success      200  {object}  utils.Response{data=SwaggerCategoryResponse} "获取成功"
// @Failure      404  {object}  utils.Response "分类不存在"
// @Router       /categories/slug/{slug} [get]
func (ctrl *CategoryController) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ValidationError(c, "无效的分类Slug")
		return
	}

	category, err := ctrl.categoryService.GetBySlug(slug)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, category)
}

// List 获取分类列表
// @Summary      获取分类列表
// @Description  分页获取分类列表
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=SwaggerCategoryListData} "获取成功"
// @Router       /categories [get]
func (ctrl *CategoryController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	categories, total, err := ctrl.categoryService.List(page, pageSize)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        categories,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAll 获取所有分类
// @Summary      获取所有分类
// @Description  获取所有分类（不分页，用于下拉选择）
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]SwaggerCategoryResponse} "获取成功"
// @Router       /categories/all [get]
func (ctrl *CategoryController) GetAll(c *gin.Context) {
	categories, err := ctrl.categoryService.GetAll()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, categories)
}
