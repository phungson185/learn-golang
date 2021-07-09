package controllers

import (
	"jwt-authen/database"
	"jwt-authen/models"
	"jwt-authen/services"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"math/rand"
)

const SecretKey = "secret"

func Register(c *gin.Context) {
	var json map[string]string
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(999999) + 100000

	password, err := bcrypt.GenerateFromPassword([]byte(json["password"]), bcrypt.DefaultCost)

	if err != nil {
		panic(err)
	}

	user := models.User{
		Email:       json["email"],
		Password:    string(password),
		Phone:       json["phone"],
		UserAddress: json["userAddress"],
		VerifyCode:  strconv.Itoa(randNum),
	}

	if err := database.Db.Create(&user); err.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not registerc"})
		return
	}

	content := "Your authentication code is: " + strconv.Itoa(randNum)
	services.SendMail(json["email"], "Verify Email", content)

	c.JSON(http.StatusOK, &user)
}

func VerifyEmail(c *gin.Context) {
	var json map[string]string
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User

	if err := database.Db.Model(&user).Where("email = ? AND verify_code = ?", json["email"], json["verifyCode"]).Update("status", true); err.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "could not verify"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification successful"})
}

func Login(c *gin.Context) {
	var json map[string]string

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.Db.Model(&user).Where("email = ? AND status = true", json["email"]).First(&user); err.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(json["password"])); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "incorrect password"})
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not login"})
		return
	}

	c.SetCookie("jwt", token, int(time.Now().Add(time.Hour*24).Unix()), "/auth", "localhost", false, true)

	c.JSON(http.StatusOK, token)
}

func Profile(c *gin.Context) {
	cookie, _ := c.Cookie("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	database.Db.Where("email = ?", claims.Issuer).First(&user)

	c.JSON(http.StatusOK, user)
}