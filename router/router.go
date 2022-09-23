package router

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/controller"
	"github.com/joexu01/container-dispatcher/docs"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Example API
// @version         1.0
// @description     This is a sample server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8880
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"https"}

	r := gin.Default()
	// set s lower memory limit for multipart forms
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	apiGroup := r.Group("/api")

	apiGroup.Use(middlewares...)
	apiGroup.Use(middleware.CORSMiddleWare())
	apiGroup.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	store := sessions.NewCookieStore([]byte("secret"))

	// User Login
	loginGroup := apiGroup.Group("")
	loginGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
	)
	{
		controller.UserControllerRegister(loginGroup)
	}

	userGroup := apiGroup.Group("/user")
	userGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.UserLogoutRegister(userGroup)
	}

	resourceGroup := apiGroup.Group("/resource")
	resourceGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.ResourceControllerRegister(resourceGroup)
	}

	imageGroup := apiGroup.Group("/image")
	imageGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.ImageControllerRegister(imageGroup)
	}

	ctnGroup := apiGroup.Group("/container")
	ctnGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.ContainerControllerRegister(ctnGroup)
	}

	algorithmGroup := apiGroup.Group("/algorithm")
	algorithmGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.AlgorithmControllerRegister(algorithmGroup)
	}

	taskGroup := apiGroup.Group("/task")
	taskGroup.Use(
		sessions.Sessions("gin-session", store),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.ValidatorBasicMiddleware(),
		middleware.SessionAuthMiddleware(),
	)
	{
		controller.TaskControllerRegister(taskGroup)
	}

	return r
}
