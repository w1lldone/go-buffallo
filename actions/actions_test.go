package actions

import (
	"coke/models"
	"encoding/json"
	"os"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/suite/v4"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}

func Login(as *ActionSuite) (token string, err error) {
	user := &models.User{
		Name:        "admin",
		Email:       "admin@mail.com",
		Password:    "password",
		AccessLevel: nulls.NewInt(4),
	}
	err = as.DB.Create(user)
	if err != nil {
		return "", err
	}
	body := &credential{
		Email:    user.Email,
		Password: "password",
	}
	res := as.JSON("/auth").Post(body)

	var m map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &m)
	if err != nil {
		return "", err
	}
	as.App.Logger.Debugf("token", m)
	token = m["token"].(string)

	return token, nil
}
