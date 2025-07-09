package routes

import (
	"github.com/Prototype-1/authentication_microservice/config"
	"github.com/Prototype-1/authentication_microservice/internal/handler"
	"github.com/Prototype-1/authentication_microservice/middleware"
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
		api.POST("/login", userHandler.Login)
		api.POST("/guest-login", userHandler.GuestLogin)
		api.POST("/verify-credentials", userHandler.VerifyCredentials)
		api.POST("/forgot-password", userHandler.ForgotPassword)
		api.POST("/create-new-password", userHandler.CreateNewPassword)
	}

	apiAuth := api.Group("/")
	apiAuth.Use(middleware.AuthMiddleware())
	{
		apiAuth.POST("/change-email-password", userHandler.ChangeEmailPassword)
		apiAuth.POST("/add-2fa", userHandler.Add2FA)
		apiAuth.POST("/add-other-credential", userHandler.AddOtherCredential)
	}
	return r
}

