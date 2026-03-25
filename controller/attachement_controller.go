// controller/attachment_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

type AttachmentController struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentController(attachmentService *service.AttachmentService) *AttachmentController {
	return &AttachmentController{attachmentService: attachmentService}
}

// Upload 上传文件
func (ctrl *AttachmentController) Upload(c *gin.Context) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.ValidationError(c, "请选择要上传的文件")
		return
	}

	// 上传文件
	attachment, err := ctrl.attachmentService.Upload(file, userID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "上传成功", attachment)
}

// UploadMultiple 多文件上传
func (ctrl *AttachmentController) UploadMultiple(c *gin.Context) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	// 获取表单中的文件列表
	form, err := c.MultipartForm()
	if err != nil {
		utils.ValidationError(c, "请选择要上传的文件")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.ValidationError(c, "请选择要上传的文件")
		return
	}

	var results []interface{}
	var errors []string

	for _, file := range files {
		attachment, err := ctrl.attachmentService.Upload(file, userID)
		if err != nil {
			errors = append(errors, file.Filename+": "+err.Error())
		} else {
			results = append(results, attachment)
		}
	}

	if len(errors) > 0 {
		utils.SuccessWithMessage(c, "部分文件上传失败", gin.H{
			"success": results,
			"errors":  errors,
		})
	} else {
		utils.SuccessWithMessage(c, "上传成功", results)
	}
}

// Delete 删除附件
func (ctrl *AttachmentController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的附件ID")
		return
	}

	// 获取当前用户信息
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	role, _ := middleware.GetUserRole(c)

	err = ctrl.attachmentService.Delete(uint(id), userID, role)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}

// GetByID 获取附件详情
func (ctrl *AttachmentController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(c, "无效的附件ID")
		return
	}

	attachment, err := ctrl.attachmentService.GetByID(uint(id))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, attachment)
}

// List 获取附件列表
func (ctrl *AttachmentController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	fileType := c.DefaultQuery("type", "")

	// 获取当前用户信息（可选认证）
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)

	// 非管理员只能查看自己的文件
	var queryUserID uint = 0
	if role != "admin" && userID > 0 {
		queryUserID = userID
	}

	attachments, total, err := ctrl.attachmentService.List(page, pageSize, fileType, queryUserID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"list":        attachments,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetUserAttachments 获取当前用户的附件
func (ctrl *AttachmentController) GetUserAttachments(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.Error(c, utils.ErrInvalidToken)
		return
	}

	attachments, err := ctrl.attachmentService.GetUserAttachments(userID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, attachments)
}
