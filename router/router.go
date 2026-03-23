// router/router.go
package router

import (
	"blog/controller"
	"blog/dao"
	"blog/models"
	"blog/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 初始化DAO
	userDAO := dao.NewUserDAO(models.DB)
	postDAO := dao.NewPostDAO(models.DB)

	// 初始化Service
	userService := service.NewUserService(userDAO)
	postService := service.NewPostService(postDAO)

	// 初始化Controller
	userController := controller.NewUserController(userService)
	postController := controller.NewPostController(postService)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 用户相关
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
		}

		// 文章相关
		posts := api.Group("/posts")
		{
			posts.GET("", postController.ListPosts)
			posts.GET("/:id", postController.GetPost)
		}

		// 需要认证的路由（后续添加JWT中间件）
		authRequired := api.Group("")
		// TODO: authRequired.Use(middleware.JWTAuth())
		{
			// 用户相关（需要认证）
			authRequired.GET("/profile", userController.GetProfile)

			// 文章管理（需要认证）
			authRequired.POST("/posts", postController.Create)           // 创建文章
			authRequired.PUT("/posts/:id", postController.UpdatePost)    // 更新文章
			authRequired.DELETE("/posts/:id", postController.DeletePost) // 删除文章
		}
	}

	return r
}
