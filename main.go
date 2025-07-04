package main

import (
	"context"
	"log"
	"os"
	"github.com/Prototype-1/authentication_microservice/config"
	"github.com/Prototype-1/authentication_microservice/routes"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	firestoreHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if firestoreHost != "" {
		os.Setenv("FIRESTORE_EMULATOR_HOST", firestoreHost)
		log.Printf("Using Firestore Emulator at %s", firestoreHost)
	}

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
