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

	// 初始化Service
	userService := service.NewUserService(userDAO, cfg)
	postService := service.NewPostService(postDAO)
	categoryService := service.NewCategoryService(categoryDAO)
	tagService := service.NewTagService(tagDAO)

	// 初始化Controller
	userController := controller.NewUserController(userService)
	postController := controller.NewPostController(postService)
	categoryController := controller.NewCategoryController(categoryService)
	tagController := controller.NewTagController(tagService)

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
		}

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

			// 文章管理
			authRequired.POST("/posts", postController.Create)
			authRequired.PUT("/posts/:id", postController.UpdatePost)
			authRequired.DELETE("/posts/:id", postController.DeletePost)
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
		}
	}

	return r
}
