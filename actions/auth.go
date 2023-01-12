package actions

import (
	"coke/internal/cache"
	"coke/models"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var MaxAttempts = 5

// AuthAuth default implementation.
func AuthCreate(c buffalo.Context) error {
	var err error
	attempts := 0

	user := &models.User{}
	err = c.Bind(user)
	if err != nil {
		return err
	}
	verr := validate.Validate(
		&validators.EmailIsPresent{Name: "email", Field: user.Email},
		&validators.StringIsPresent{Name: "password", Field: user.Password},
	)
	if verr.HasAny() {
		response := Response{
			Errors: verr.Errors,
			Status: "error",
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(response))
	}

	res, err := cache.Cache.Value(getAttemptsCacheKey(user.Email))
	if err == nil {
		attempts = res.Data().(int) + attempts
	}
	if attempts >= MaxAttempts {
		return c.Render(http.StatusTooManyRequests, r.JSON(Response{
			Errors: fmt.Sprintf("Too many attempts. Please try again in %v minutes", math.Floor(res.LifeSpan().Minutes())),
		}))
	}

	password := user.Password
	err = models.DB.Where("email = (?)", user.Email).First(user)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		cache.Cache.Add(getAttemptsCacheKey(user.Email), 5*time.Minute, attempts+1)
		return err
	}

	claims := jwt.MapClaims{}
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(7 * 24 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(envy.Get("JWT_SECRET", "")))
	if err != nil {
		return err
	}

	cache.Cache.Delete(getAttemptsCacheKey(user.Email))

	response := make(map[string]string)
	response["token"] = tokenString
	return c.Render(http.StatusOK, r.JSON(response))
}

func AuthIndex(c buffalo.Context) error {
	auth := c.Value("auth").(*models.User)
	response := Response{
		Data: auth,
	}
	return c.Render(http.StatusOK, r.JSON(response))
}

func getAttemptsCacheKey(email string) string {
	return fmt.Sprintf("attempt:%s", email)
}
