// controller/category_controller.go
package controller

import (
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

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

// Create 创建分类（需要管理员权限）
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

// Update 更新分类（需要管理员权限）
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

// Delete 删除分类（需要管理员权限）
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

// GetByID 获取分类详情（公开）
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

// GetBySlug 通过Slug获取分类详情（公开）
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

// List 获取分类列表（公开）
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

// GetAll 获取所有分类（不分页，用于下拉选择）
func (ctrl *CategoryController) GetAll(c *gin.Context) {
	categories, err := ctrl.categoryService.GetAll()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, categories)
}
