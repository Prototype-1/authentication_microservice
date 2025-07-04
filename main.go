package main

import (
	"context"
	"log"
	"github.com/Prototype-1/authentication_microservice/config"
	"github.com/Prototype-1/authentication_microservice/routes"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	opt := option.WithCredentialsFile(cfg.FirebaseCredFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error initializing Firebase Auth: %v", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing Firestore: %v", err)
	}
	defer firestoreClient.Close()
	r := routes.SetupRouter(cfg, authClient, firestoreClient)

	log.Printf("Server running on port %s", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}
