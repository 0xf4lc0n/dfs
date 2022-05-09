package controllers

import (
	"dfs/auth/di"
	"dfs/auth/dtos"
	"dfs/auth/models"
	"dfs/auth/validation"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const SecretKey = "secret"

func Register(c *fiber.Ctx) error {
	registerDto := new(dtos.RegisterDto)

	if err := c.BodyParser(&registerDto); err != nil {
		di.Logger.Warn("Cannot parse register data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var user models.User

	if err := di.Db.Where("email = ?", registerDto.Email).First(&user).Error; err == nil {
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

	if err := di.MailService.SendMail(registerDto.Name, registerDto.Email, verificationData.Code); err != nil {
		di.Logger.Error("Cannot send mail", zap.Error(err))
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Cannot send verification mail on the given email address",
		})
	}

	di.Db.Create(&verificationData)

	password, _ := bcrypt.GenerateFromPassword([]byte(registerDto.Password), 14)

	user = models.User{
		Name:     registerDto.Name,
		Email:    registerDto.Email,
		Password: password,
		Verified: false,
	}

	di.Db.Create(&user)

	return c.SendStatus(fiber.StatusCreated)
}

func Login(c *fiber.Ctx) error {
	loginDto := new(dtos.LoginDto)

	if err := c.BodyParser(&loginDto); err != nil {
		di.Logger.Warn("Cannot parse login data.", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	errors := validation.Validate(loginDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	var user models.User

	di.Db.Where("email = ?", loginDto.Email).First(&user)

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

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	di.Db.Where("id = ?", claims.Issuer).First(&user)

	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.SendStatus(fiber.StatusOK)
}

func VerifyEmail(c *fiber.Ctx) error {
	code := c.Params("code")

	var verificationData models.VerificationData

	if err := di.Db.Where("code = ?", code).First(&verificationData).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	var user models.User

	if time.Now().After(verificationData.ExpiresAt) {
		di.Db.Delete(&verificationData)
		di.Db.Where("email = ?", verificationData.Email).Delete(&user)
		return c.SendStatus(fiber.StatusNotFound)
	}

	if err := di.Db.Where("email = ?", verificationData.Email).First(&user).Error; err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	di.Db.Model(&user).Update("verified", true)
	di.Db.Delete(&verificationData)

	return c.SendStatus(fiber.StatusOK)
}
