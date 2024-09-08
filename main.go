package main

import (
	"encoding/base64"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	secretKey = []byte("my_secret_key")
	db        *gorm.DB
)

type Payload struct {
	GUID string
	IP   string
	jwt.StandardClaims
}

type User struct {
	GUID         string `gorm:"primaryKey"`
	IP           string
	Email        string
	RefreshToken string
}

func GenerateAccessToken(GUID string, IP string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Payload{
		GUID: GUID,
		IP:   IP,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	accessToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GenerateRefreshToken() (string, error) {
	return uuid.New().String(), nil
}

func CreateUser(GUID string, IP string, refreshToken string) {
	email := "test@mail.ru"
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Ошибка шифрования refresh токена: %v", err)
	}

	user := &User{
		GUID:         GUID,
		IP:           IP,
		Email:        email,
		RefreshToken: string(hashedRefreshToken),
	}

	db.Create(&user)
}

func HandleTokensGeneration(c *gin.Context) {
	GUID := c.Query("guid") // TODO: 400, если не переданы параметры в запрос
	IP := c.ClientIP()
	accessToken, err := GenerateAccessToken(GUID, IP)
	if err != nil {
		log.Fatalf("Ошибка генерации access токена: %v", err)
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		log.Fatalf("Ошибка генерации refresh токена: %v", err)
	}

	CreateUser(GUID, IP, refreshToken)

	c.JSON(200, gin.H{
		"access_token":  accessToken,
		"refresh_token": base64.StdEncoding.EncodeToString([]byte(refreshToken)),
	})
}

func main() {
	var err error
	dsn := "host=db user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Ошибка подключения к бд")
	}

	db.AutoMigrate(&User{})

	router := gin.Default()
	router.GET("/generate", HandleTokensGeneration)

	router.Run("0.0.0.0:8080")
}
