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
			posts.POST("", postController.Create) // TODO: 添加JWT中间件
		}
	}

	return r
}
