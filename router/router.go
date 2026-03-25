// router/router.go
package router

import (
	"blog/config"
	"blog/controller"
	"blog/dao"
	"blog/middleware"
	"blog/models"
	"blog/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// 初始化DAO
	userDAO := dao.NewUserDAO(models.DB)
	postDAO := dao.NewPostDAO(models.DB)
	categoryDAO := dao.NewCategoryDAO(models.DB)
	tagDAO := dao.NewTagDAO(models.DB)
	commentDAO := dao.NewCommentDAO(models.DB)
	attachmentDAO := dao.NewAttachmentDAO(models.DB)

	// 初始化Service
	userService := service.NewUserService(userDAO, cfg)
	postService := service.NewPostService(postDAO, tagDAO, categoryDAO, models.DB)
	categoryService := service.NewCategoryService(categoryDAO)
	tagService := service.NewTagService(tagDAO)
	commentService := service.NewCommentService(commentDAO, postDAO, userDAO, models.DB)
	attachmentService := service.NewAttachmentService(attachmentDAO, "./uploads", "http://localhost:8080")

	// 初始化Controller
	userController := controller.NewUserController(userService)
	postController := controller.NewPostController(postService)
	categoryController := controller.NewCategoryController(categoryService)
	tagController := controller.NewTagController(tagService)
	commentController := controller.NewCommentController(commentService)
	attachmentController := controller.NewAttachmentController(attachmentService) // 新增

	// API路由组
	api := r.Group("/api/v1")
	{
		// 公开路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
			auth.POST("/refresh", userController.RefreshToken)

		}

		// 分类公开路由
		categories := api.Group("/categories")
		{
			categories.GET("", categoryController.List)                 // 获取分类列表
			categories.GET("/all", categoryController.GetAll)           // 获取所有分类
			categories.GET("/:id", categoryController.GetByID)          // 获取分类详情
			categories.GET("/slug/:slug", categoryController.GetBySlug) // 通过slug获取
		}

		// 文章公开路由
		posts := api.Group("/posts")
		{
			posts.GET("", postController.ListPosts)
			posts.GET("/:id", postController.GetPost)
			posts.GET("/tag/:slug", postController.GetPostsByTag)
		}

		api.GET("/posts/:id/comments", commentController.GetPostComments)

		// 标签公开路由
		tags := api.Group("/tags")
		{
			tags.GET("", tagController.List)                 // 标签列表
			tags.GET("/all", tagController.GetAll)           // 所有标签
			tags.GET("/popular", tagController.GetPopular)   // 热门标签
			tags.GET("/:id", tagController.GetByID)          // 标签详情
			tags.GET("/slug/:slug", tagController.GetBySlug) // 通过slug获取
		}

		// 需要认证的路由
		authRequired := api.Group("")
		authRequired.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 用户相关
			authRequired.GET("/profile", userController.GetProfile)
			authRequired.GET("/profile/comments", commentController.GetUserComments) // 我的评论

			// 文章管理
			authRequired.POST("/posts", postController.Create)
			authRequired.PUT("/posts/:id", postController.UpdatePost)
			authRequired.DELETE("/posts/:id", postController.DeletePost)

			// 评论操作
			authRequired.POST("/posts/:id/comments", commentController.CreateComment)   // 发表评论
			authRequired.DELETE("/comments/:cid", commentController.DeleteComment)      // 删除评论
			authRequired.POST("/comments/:cid/like", commentController.LikeComment)     // 点赞
			authRequired.DELETE("/comments/:cid/like", commentController.UnlikeComment) // 取消点赞

			// 文件上传
			authRequired.POST("/upload", attachmentController.Upload)
			authRequired.POST("/upload/multiple", attachmentController.UploadMultiple)
			authRequired.DELETE("/attachments/:id", attachmentController.Delete)
			authRequired.GET("/attachments", attachmentController.List)
			authRequired.GET("/attachments/my", attachmentController.GetUserAttachments)
			authRequired.GET("/attachments/:id", attachmentController.GetByID)
		}

		// 管理员专用路由（可选）
		adminRequired := api.Group("/admin")
		adminRequired.Use(middleware.JWTAuth(cfg.JWT.Secret))
		adminRequired.Use(middleware.RequireRole("admin"))
		{
			// 分类管理
			adminRequired.POST("/categories", categoryController.Create)
			adminRequired.PUT("/categories/:id", categoryController.Update)
			adminRequired.DELETE("/categories/:id", categoryController.Delete)

			// 标签管理
			adminRequired.POST("/tags", tagController.Create)
			adminRequired.PUT("/tags/:id", tagController.Update)
			adminRequired.DELETE("/tags/:id", tagController.Delete)

			// 评论管理
			adminRequired.GET("/comments/pending", commentController.GetPendingComments)   // 待审核列表
			adminRequired.POST("/comments/:cid/approve", commentController.ApproveComment) // 审核通过
			adminRequired.POST("/comments/:cid/reject", commentController.RejectComment)   // 拒绝

			// 管理员可以查看所有附件
			adminRequired.GET("/attachments/all", attachmentController.List)
		}
	}

	return r
}
