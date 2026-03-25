// controller/attachment_controller.go
package controller

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

// SwaggerAttachmentResponse 附件响应（用于 Swagger 文档）
type SwaggerAttachmentResponse struct {
	ID           uint   `json:"id" example:"1"`
	Filename     string `json:"filename" example:"550e8400-e29b-41d4-a716-446655440000.jpg"`
	OriginalName string `json:"original_name" example:"test.jpg"`
	URL          string `json:"url" example:"http://localhost:8080/uploads/2024/01/test.jpg"`
	ThumbURL     string `json:"thumb_url,omitempty" example:"http://localhost:8080/uploads/2024/01/thumb_test.jpg"`
	Size         int64  `json:"size" example:"102400"`
	MimeType     string `json:"mime_type" example:"image/jpeg"`
	Type         string `json:"type" example:"image"`
	Width        int    `json:"width,omitempty" example:"800"`
	Height       int    `json:"height,omitempty" example:"600"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// SwaggerAttachmentListData 附件列表响应（用于 Swagger 文档）
type SwaggerAttachmentListData struct {
	List       []SwaggerAttachmentResponse `json:"list"`
	Total      int64                       `json:"total" example:"100"`
	Page       int                         `json:"page" example:"1"`
	PageSize   int                         `json:"page_size" example:"10"`
	TotalPages int64                       `json:"total_pages" example:"10"`
}

// AttachmentResponse 附件响应

type AttachmentController struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentController(attachmentService *service.AttachmentService) *AttachmentController {
	return &AttachmentController{attachmentService: attachmentService}
}

// Upload 上传文件
// @Summary      上传文件
// @Description  上传单个文件（图片、文档等）（需要认证）
// @Tags         文件管理
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "要上传的文件"
// @Success      200  {object}  utils.Response{data=SwaggerAttachmentResponse} "上传成功"
// @Failure      400  {object}  utils.Response "参数错误或文件过大"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /upload [post]
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
// @Summary      多文件上传
// @Description  同时上传多个文件（需要认证）
// @Tags         文件管理
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        files formData file true "要上传的文件列表"
// @Success      200  {object}  utils.Response{data=object{success=[]SwaggerAttachmentResponse,errors=[]string}} "上传成功（部分失败时返回错误列表）"
// @Failure      400  {object}  utils.Response "参数错误"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /upload/multiple [post]
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
// @Summary      删除附件
// @Description  删除附件（需要认证，只能删除自己的文件）
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "附件ID"
// @Success      200  {object}  utils.Response "删除成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Failure      403  {object}  utils.Response "权限不足"
// @Failure      404  {object}  utils.Response "附件不存在"
// @Router       /attachments/{id} [delete]
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
// @Summary      获取附件详情
// @Description  根据ID获取附件详情
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "附件ID"
// @Success      200  {object}  utils.Response{data=SwaggerAttachmentResponse} "获取成功"
// @Failure      404  {object}  utils.Response "附件不存在"
// @Router       /attachments/{id} [get]
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
// @Summary      获取附件列表
// @Description  分页获取附件列表（用户只能看到自己的文件）
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Param        type query string false "文件类型 image/file/video"
// @Success      200  {object}  utils.Response{data=SwaggerAttachmentListData} "获取成功"
// @Router       /attachments [get]
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
// @Summary      获取当前用户的附件
// @Description  获取当前登录用户的所有附件（需要认证）
// @Tags         文件管理
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=[]SwaggerAttachmentResponse} "获取成功"
// @Failure      401  {object}  utils.Response "未授权"
// @Router       /attachments/my [get]
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
