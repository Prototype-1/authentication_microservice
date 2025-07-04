package handler

import (
	"context"
	"log"
	"net/http"
	"time"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4/auth"
	"github.com/Prototype-1/authentication_microservice/internal/model"
	"github.com/Prototype-1/authentication_microservice/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	authClient      *firebase.Client
	firestoreClient *firestore.Client
}

func NewUserHandler(auth *firebase.Client, fs *firestore.Client) *UserHandler {
	return &UserHandler{
		authClient:      auth,
		firestoreClient: fs,
	}
}

type SignUpRequest struct {
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password" binding:"required"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Country      string `json:"country"`
	Address      string `json:"address"`
}

func (h *UserHandler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Email == "" && req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or phone number is required"})
		return
	}

	if req.Email != "" {
	log.Print("Please provide a valid email")
	}

	if req.PhoneNumber != "" && !service.ValidatePhoneNumber(req.PhoneNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number format. should be in +countrycode format"})
		return
	}

	ctx := context.Background()

	params := (&firebase.UserToCreate{}).
		Email(req.Email).
		PhoneNumber(req.PhoneNumber).
		Password(req.Password).
		EmailVerified(false)

	u, err := h.authClient.CreateUser(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user in firebase: " + err.Error()})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := model.User{
		UID:                 u.UID,
		Email:               req.Email,
		PhoneNumber:         req.PhoneNumber,
		IsPhoneVerified:     false,
		IsEmailVerified:     false,
		IsGuestUser:         false,
		Password:            string(hashedPwd),
		Joint:               []string{"Capcons"},
		IsBillableUser:      false,
		Is2FNeeded:          false,
		UserFirstName:       req.FirstName,
		UserSecondName:      req.LastName,
		UserCreatedDate:     time.Now(),
		UserLastLoginDetail: time.Now(),
		CountryOfOrigin:     req.Country,
		Address:             req.Address,
	}

	_, err = h.firestoreClient.Collection("users").Doc(u.UID).Set(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store user in Firestore: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "signup successful", "uid": u.UID})
}
