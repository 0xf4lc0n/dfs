package controllers

import (
	"crypto/rand"
	"dfs/auth/dtos"
	"dfs/auth/models"
	"dfs/auth/services"
	"dfs/auth/validation"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const SecretKey = "secret"

type AuthController struct {
	logger    *zap.Logger
	database  *gorm.DB
	mail      *services.MailService
	rpcClient *services.RpcClient
}

func NewAuthController(logger *zap.Logger, database *gorm.DB, mail *services.MailService, rpcClient *services.RpcClient) *AuthController {
	return &AuthController{logger: logger, database: database, mail: mail, rpcClient: rpcClient}
}

func (ac *AuthController) RegisterRoutes(app *fiber.App) {
	app.Post("/api/register", ac.Register)
	app.Post("/api/login", ac.Login)
	app.Post("/api/logout", ac.Logout)
	app.Get("/api/user", ac.User)
	app.Get("/api/verify/:code", ac.VerifyEmail)
}

func (ac *AuthController) Register(c *fiber.Ctx) error {
	registerDto := new(dtos.RegisterDto)

	if err := c.BodyParser(&registerDto); err != nil {
		ac.logger.Warn("Cannot parse register data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var user models.User

	if err := ac.database.Where("email = ?", registerDto.Email).First(&user).Error; err == nil {
		c.Status(fiber.StatusConflict)
		return c.JSON(fiber.Map{
			"message": "This email address is already taken",
		})
	}

	errors := validation.Validate(registerDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	verificationData := models.VerificationData{
		Email:     registerDto.Email,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		Code:      strings.Replace(uuid.New().String(), "-", "", -1),
	}

	if err := ac.mail.SendMail(registerDto.Name, registerDto.Email, verificationData.Code); err != nil {
		ac.logger.Error("Cannot send mail", zap.Error(err))
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Cannot send verification mail on the given email address",
		})
	}

	ac.database.Create(&verificationData)

	password, _ := bcrypt.GenerateFromPassword([]byte(registerDto.Password), 14)

	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		ac.logger.Error("Cannot generate encryption key")
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Cannot send verification mail on the given email address",
		})
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)

	user = models.User{
		Name:          registerDto.Name,
		Email:         registerDto.Email,
		Password:      password,
		Verified:      false,
		HomeDirectory: registerDto.Email,
		CryptKey:      encodedKey,
	}

	if ac.rpcClient.CreateHomeDirectory(user.HomeDirectory) == false {
		ac.logger.Error("Cannot create home directory for user:", zap.String("User", user.Email))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot create account",
		})
	}

	ac.database.Create(&user)

	return c.SendStatus(fiber.StatusCreated)
}

func (ac *AuthController) Login(c *fiber.Ctx) error {
	loginDto := new(dtos.LoginDto)

	if err := c.BodyParser(&loginDto); err != nil {
		ac.logger.Warn("Cannot parse login data.", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	errors := validation.Validate(loginDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	var user models.User

	ac.database.Where("email = ?", loginDto.Email).First(&user)

	if user.Id == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Incorrect login or password",
		})
	}

	if user.Verified == false {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "You have to verify your email address",
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(loginDto.Password)); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Incorrect login or password",
		})
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Could not login",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.SendStatus(fiber.StatusOK)
}

func (ac *AuthController) User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		ac.logger.Info("", zap.Error(err))
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	ac.database.Where("id = ?", claims.Issuer).First(&user)

	return c.JSON(user)
}

func (ac *AuthController) Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.SendStatus(fiber.StatusOK)
}

func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	code := c.Params("code")

	var verificationData models.VerificationData

	if err := ac.database.Where("code = ?", code).First(&verificationData).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	var user models.User

	if time.Now().After(verificationData.ExpiresAt) {
		ac.database.Delete(&verificationData)
		ac.database.Where("email = ?", verificationData.Email).Delete(&user)
		return c.SendStatus(fiber.StatusNotFound)
	}

	if err := ac.database.Where("email = ?", verificationData.Email).First(&user).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	ac.database.Model(&user).Update("verified", true)
	ac.database.Delete(&verificationData)

	return c.SendStatus(fiber.StatusOK)
}
