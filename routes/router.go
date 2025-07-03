package routes

import (
	"github.com/Prototype-1/authentication_microservice/config"
	"github.com/Prototype-1/authentication_microservice/internal/handler"
	"github.com/gin-gonic/gin"
	firebase "firebase.google.com/go/v4/auth"
	"cloud.google.com/go/firestore"
)

func SetupRouter(cfg *config.Config, authClient *firebase.Client, firestoreClient *firestore.Client) *gin.Engine {
	r := gin.Default()
	userHandler := handler.NewUserHandler(authClient, firestoreClient)

	api := r.Group("/api")
	{
		api.POST("/signup", userHandler.SignUp)

	}

	return r
}
