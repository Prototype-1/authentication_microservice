# Authentication Microservice

This is a Go (Golang) microservice that implements robust authentication using Firebase Auth and Firestore.  
It was built as part of a machine task to demonstrate secure authentication workflows, clean architecture, and integrations with external data stores.

---

## Features

 User signup with email / phone and strong password validation  
 Secure password storage using bcrypt  
 Guest user creation (anonymous login)  
 Login endpoint that verifies hashed password  
 Forgot password with secure Firebase reset link  
 Ability to create a new password directly  
 Change email and/or password  
 Toggle 2FA requirement flag  
 Add additional credentials (phone or email if the other is present)  
 Firestore triggers (outside scope) handle syncing data to Neo4J and MongoDB

---

## Technologies Used

-  **Go (Golang)** — Gin framework for REST APIs
-  **Firebase Auth** — user identity, password resets
-  **Firestore** — extended user profiles and flags
-  **bcrypt** — secure password hashing
-  **godotenv** — load environment variables
-  **validator/v10** — struct validation for inputs
-  Clean layered architecture
    - `main.go` - bootstraps server
    - `config/` - env + Firebase/Firestore setup
    - `routes/` - Gin routes setup
    - `internal/` - domain logic
        - `handler` - HTTP handlers
        - `service` - validation helpers etc.
        - `model` - data models
    - `.env` - all config secrets

---

## Environment Variables

`.env` file:
```env
MONGO_URI=mongodb://mongodb:27017
MONGO_DB=authentication_service

SERVER_PORT=8081

FIREBASE_CREDENTIALS=auth-task-project-firebase-adminsdk-xxxx.json
```

## Running the service

go mod tidy
go run main.go

## Server starts on:

http://localhost:8081

## Notes

    Passwords are bcrypt hashed before storing in Firestore.

    Firestore triggers (assumed to be set up externally) handle syncing user data into Neo4J and MongoDB.

    Each endpoint is kept RESTful, following clear input/output JSON contracts.

    No body required for /guest-login.

## Future improvements

    Swagger/OpenAPI for live docs

    Docker + docker-compose

    JWT sessions

    Role-based access with middleware