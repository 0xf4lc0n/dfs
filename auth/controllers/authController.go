package controllers

import (
	"crypto/rand"
	"dfs/auth/database"
	"dfs/auth/dtos"
	"dfs/auth/services"
	"dfs/auth/validation"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const SecretKey = "secret"

type AuthController struct {
	logger   *zap.Logger
	userRepo *database.UserRepository
	vrfRepo  *database.VerificationRepository
	emailSrv *services.MailService
	rpc      *services.RpcClient
}

func NewAuthController(logger *zap.Logger, usrRepo *database.UserRepository, vrfRepo *database.VerificationRepository,
	mail *services.MailService, rpcClient *services.RpcClient) *AuthController {
	return &AuthController{logger: logger, userRepo: usrRepo, vrfRepo: vrfRepo, emailSrv: mail, rpc: rpcClient}
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

	user := ac.userRepo.GetUserByEmail(registerDto.Email)

	if user != nil {
		return c.JSON(fiber.Map{
			"message": "This email address is already taken",
		})
	}

	errors := validation.Validate(registerDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	verificationData := ac.vrfRepo.CreateAndReturnVerificationData(registerDto.Email)

	if verificationData == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Cannot create account"})
	}

	if err := ac.emailSrv.SendMail(registerDto.Name, registerDto.Email, verificationData.Code); err != nil {
		ac.vrfRepo.DeleteVerification(verificationData.Id)
		ac.logger.Error("Cannot send verification mail", zap.Error(err))
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "Cannot send verification mail on the given email address",
		})
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(registerDto.Password), 14)

	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		ac.logger.Error("Cannot generate encryption key")
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "Cannot create account",
		})
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)

	if ac.rpc.CreateHomeDirectory(registerDto.Email) == false {
		ac.logger.Error("Cannot create home directory for user:", zap.String("User", user.Email))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Cannot create account",
		})
	}

	if ac.userRepo.CreateUser(registerDto.Name, registerDto.Email, password, encodedKey) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Cannot create account"})
	}

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

	user := ac.userRepo.GetUserByEmail(loginDto.Email)

	if user == nil {
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
		ac.logger.Error("Cannot parse JWT claims", zap.Error(err))
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*jwt.StandardClaims)

	userId, err := strconv.ParseUint(claims.Issuer, 10, 0)

	if err != nil {
		ac.logger.Error("Cannot convert claims Issuer to uint", zap.Error(err))
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	user := ac.userRepo.GetUserById(uint(userId))

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

	verificationData := ac.vrfRepo.GetVerificationByCode(code)

	if verificationData == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if time.Now().After(verificationData.ExpiresAt) {
		ac.vrfRepo.DeleteVerification(verificationData.Id)
		ac.userRepo.DeleteUserByEmail(verificationData.Email)
		return c.SendStatus(fiber.StatusNotFound)
	}

	user := ac.userRepo.GetUserByEmail(verificationData.Email)

	if user == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	if ac.userRepo.VerifyUser(user) == false {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	ac.vrfRepo.DeleteVerification(verificationData.Id)

	return c.SendStatus(fiber.StatusOK)
}
