// @title       Voice Provisioning API
// @description Documentation of Voice API
// @schemes     http
// @BasePath    /api/v1
// @version     1.0.0
// @host        localhost:11010
// @accept      json
// @produce     json
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Openline-API-KEY

package routes

import (
	"github.com/gin-gonic/gin"
	c "github.com/openline-ai/openline-voice/packages/apps/voice-plugin/config"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/docs"
	"github.com/openline-ai/openline-voice/packages/apps/voice-plugin/gen"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
)

// Run will start the server
func Run(conf *c.Config, client *gen.Client) {
	router := getRouter(conf, client)
	if err := router.Run(conf.Service.ServerAddress); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func checkAPIKeyMiddleware(conf *c.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Openline-API-KEY")

		// Check if the API key is valid
		if apiKey != conf.Service.ApiKey {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}

func getRouter(conf *c.Config, client *gen.Client) *gin.Engine {
	router := gin.New()
	route := router.Group("/api/v1")
	route.Use(checkAPIKeyMiddleware(conf))
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = conf.Service.Host
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	log.Printf("Got a host of %s", conf.Service.Host)

	addAddressRoutes(conf, client, route)
	addNumberRoutes(conf, client, route)
	addCarrierRoutes(conf, client, route)
	addProfileRoutes(conf, client, route)
	addSipEndpointRoutes(conf, client, route)
	addVoiceMailRoutes(conf, client, route)
	addForwardingRoutes(conf, client, route)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}
