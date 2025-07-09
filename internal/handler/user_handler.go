package handler

import (
	"context"
	"net/http"
	"time"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4/auth"
	"github.com/Prototype-1/authentication_microservice/internal/model"
	"github.com/Prototype-1/authentication_microservice/internal/service"
	"github.com/Prototype-1/authentication_microservice/pkg"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
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
	if _, err := mail.ParseAddress(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}
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

type LoginRequest struct {
	EmailOrPhone string `json:"email_or_phone" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	query := h.firestoreClient.Collection("users").Where("Email", "==", req.EmailOrPhone)
	if service.ValidatePhoneNumber(req.EmailOrPhone) {
		query = h.firestoreClient.Collection("users").Where("PhoneNumber", "==", req.EmailOrPhone)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil || len(docs) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	var user model.User
	docs[0].DataTo(&user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	_, _ = h.firestoreClient.Collection("users").Doc(user.UID).Update(ctx, []firestore.Update{
		{Path: "UserLastLoginDetail", Value: time.Now()},
	})

	token, err := pkg.GenerateJWT(user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"uid":     user.UID,
		"token":   token,
	})
}

func (h *UserHandler) GuestLogin(c *gin.Context) {
	ctx := context.Background()

	u, err := h.authClient.CreateUser(ctx, &firebase.UserToCreate{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create guest user"})
		return
	}

	user := model.User{
		UID:                 u.UID,
		IsGuestUser:         true,
		IsPhoneVerified:     false,
		IsEmailVerified:     false,
		Joint:               []string{"Capcons"},
		IsBillableUser:      false,
		Is2FNeeded:          false,
		UserCreatedDate:     time.Now(),
		UserLastLoginDetail: time.Now(),
	}

	_, err = h.firestoreClient.Collection("users").Doc(u.UID).Set(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store guest in Firestore"})
		return
	}

		token, err := pkg.GenerateJWT(user.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "guest user created",
		"uid":     u.UID,
		"token": token,
	})
}

type VerifyRequest struct {
	EmailOrPhone string `json:"email_or_phone" binding:"required"`
}

func (h *UserHandler) VerifyCredentials(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	query := h.firestoreClient.Collection("users").Where("Email", "==", req.EmailOrPhone)
	if service.ValidatePhoneNumber(req.EmailOrPhone) {
		query = h.firestoreClient.Collection("users").Where("PhoneNumber", "==", req.EmailOrPhone)
	}

	docs, _ := query.Documents(ctx).GetAll()
	if len(docs) == 0 {
		c.JSON(http.StatusOK, gin.H{"exists": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": true, "uid": docs[0].Ref.ID})
}

type ForgotRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.authClient.PasswordResetLink(c, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create password reset link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reset link created",
		"link":    link,
	})
}

type NewPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) CreateNewPassword(c *gin.Context) {
	var req NewPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	u, err := h.authClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	_, err = h.authClient.UpdateUser(ctx, u.UID, (&firebase.UserToUpdate{}).Password(req.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password in firebase"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, _ = h.firestoreClient.Collection("users").Doc(u.UID).Update(ctx, []firestore.Update{
		{Path: "Password", Value: string(hashedPwd)},
	})

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

type ChangeRequest struct {
	UID      string `json:"uid" binding:"required"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) ChangeEmailPassword(c *gin.Context) {
	uidFromJWT, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uid not found in context"})
		return
	}
	uid := uidFromJWT.(string)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	params := &firebase.UserToUpdate{}
	if req.Email != "" {
		params.Email(req.Email)
	}
	if req.Password != "" {
		if err := service.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params.Password(req.Password)
	}

	_, err := h.authClient.UpdateUser(ctx, uid, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed update in firebase"})
		return
	}

	if req.Password != "" {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		_, _ = h.firestoreClient.Collection("users").Doc(uid).Update(ctx, []firestore.Update{
			{Path: "Password", Value: string(hashedPwd)},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "your credentials has been successfully updated"})
}

type TwoFARequest struct {
	UID       string `json:"uid" binding:"required"`
	Is2FNeeded bool  `json:"is2FNeeded"`
}

func (h *UserHandler) Add2FA(c *gin.Context) {
	uidFromJWT, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uid not found in context"})
		return
	}
	uid := uidFromJWT.(string)

	var req struct {
		Is2FNeeded bool `json:"is2FNeeded"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	_, _ = h.firestoreClient.Collection("users").Doc(uid).Update(ctx, []firestore.Update{
		{Path: "Is2FNeeded", Value: req.Is2FNeeded},
	})

	c.JSON(http.StatusOK, gin.H{"message": "2FA setting updated"})
}

func (h *UserHandler) AddOtherCredential(c *gin.Context) {
	uidFromJWT, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "uid not found in context"})
		return
	}
	uid := uidFromJWT.(string)

	var req struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	fields := []firestore.Update{}
	if req.Email != "" {
		fields = append(fields, firestore.Update{Path: "Email", Value: req.Email})
	}
	if req.Phone != "" {
		if !service.ValidatePhoneNumber(req.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone"})
			return
		}
		fields = append(fields, firestore.Update{Path: "PhoneNumber", Value: req.Phone})
	}
	if len(fields) > 0 {
		_, _ = h.firestoreClient.Collection("users").Doc(uid).Update(ctx, fields)
	}
	c.JSON(http.StatusOK, gin.H{"message": "credentials updated"})
}




