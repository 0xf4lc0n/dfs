package controllers

import (
	"auth/dtos"
	"auth/models"
	"auth/services"
	"auth/validation"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const SecretKey = "secret"

func Register(c *fiber.Ctx) error {
	registerDto := new(dtos.RegisterDto)

	if err := c.BodyParser(&registerDto); err != nil {
		services.Logger.Warn("Cannot parse register data", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var user models.User

	if err := services.Db.Where("email = ?", registerDto.Email).First(&user).Error; err == nil {
		c.Status(fiber.StatusConflict)
		return c.JSON(fiber.Map{
			"message": "This email address is already taken",
		})
	}

	errors := validation.Validate(registerDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(registerDto.Password), 14)

	user = models.User{
		Name:     registerDto.Name,
		Email:    registerDto.Email,
		Password: password,
	}

	services.Db.Create(&user)

	return c.SendStatus(fiber.StatusCreated)
}

func Login(c *fiber.Ctx) error {
	loginDto := new(dtos.LoginDto)

	if err := c.BodyParser(&loginDto); err != nil {
		services.Logger.Warn("Cannot parse login data.", zap.Error(err))
		return c.SendStatus(fiber.StatusBadRequest)
	}

	errors := validation.Validate(loginDto)

	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	var user models.User

	services.Db.Where("email = ?", loginDto.Email).First(&user)

	if user.Id == 0 {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Incorrect login or password",
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

	services.Db.Where("id = ?", claims.Issuer).First(&user)

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
