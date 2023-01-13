package actions

import (
	"coke/internal/cache"
	"coke/models"
	"os"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/suite/v4"
	"github.com/golang-jwt/jwt/v4"
)

type ActionSuite struct {
	*suite.Action
}

var UserAdmin *models.User

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}

	cache.NewCache(as.App.Name)

	suite.Run(t, as)
}

func NewAdmin(as *ActionSuite) error {
	user := &models.User{
		Name:        "admin",
		Email:       "admin@mail.com",
		Password:    "password",
		AccessLevel: nulls.NewInt(4),
	}

	err := as.DB.Create(user)
	if err != nil {
		return err
	}

	UserAdmin = user

	return nil
}

func Login(as *ActionSuite) (ts string, err error) {
	err = NewAdmin(as)
	if err != nil {
		as.T().Fatal("failed creating new User Admin")
	}

	claims := jwt.MapClaims{}
	claims["user_id"] = UserAdmin.ID
	claims["exp"] = time.Now().Add(7 * 24 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(envy.Get("JWT_SECRET", "")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
