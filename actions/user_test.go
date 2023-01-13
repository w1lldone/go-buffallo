package actions

import (
	"coke/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_Users_Index_Unauthorized() {
	res := as.JSON("/users").Get()

	as.Equal(http.StatusUnauthorized, res.Result().StatusCode)
}

func (as *ActionSuite) Test_Users_Index_Authorized() {
	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	req := as.JSON("/users")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	res := req.Get()
	as.Equal(http.StatusOK, res.Result().StatusCode)

	var data map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &data)
	if err != nil {
		as.Fail("unmarshal failed")
	}

	as.App.Logger.Debugf("marshalled json", data)
	users := data["data"].([]interface{})
	as.Equal(1, len(users))
}

func (as *ActionSuite) Test_Users_Show() {
	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	user := &models.User{
		Name:     "user",
		Email:    "email@mail.com",
		Password: "password",
	}
	err = as.DB.Create(user)
	if err != nil {
		as.T().Fatal("error creating user")
	}

	req := as.JSON(fmt.Sprintf("/users/%d", user.ID))
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	res := req.Get()
	as.Equal(http.StatusOK, res.Result().StatusCode)

	var data map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &data)
	if err != nil {
		as.Fail("unmarshal failed")
	}

	as.App.Logger.Debugf("Data ", data["data"].(map[string]interface{}))
	dataUser := data["data"].(map[string]interface{})
	as.Equal(dataUser["id"], float64(user.ID))
}

func (as *ActionSuite) Test_Users_Create() {
	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	user := userJson{
		Name:                 "user",
		Email:                "user@mail.com",
		Password:             "password",
		PasswordConfirmation: "password",
		AccessLevel:          nulls.NewInt(2),
	}
	req := as.JSON("/users")
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	res := req.Post(user)

	var resJson map[string]interface{}
	err = json.Unmarshal(res.Body.Bytes(), &resJson)
	if err != nil {
		as.Fail("Failed marshal json")
	}

	as.App.Logger.Debugf("json response", resJson)
	as.Equal(http.StatusCreated, res.Result().StatusCode)

	count, err := as.DB.Where("email = ?", user.Email).Count(&models.User{})
	if err != nil {
		as.FailNow("error counting results", err)
	}

	as.Equal(1, count)
}

func (as *ActionSuite) Test_Users_Update() {
	user := &models.User{
		Name:     "user",
		Email:    "email@mail.com",
		Password: "password",
	}
	err := as.DB.Create(user)
	if err != nil {
		as.T().Fatal("error creating user")
	}

	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	req := as.JSON(fmt.Sprintf("/users/%d", user.ID))
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	body := &models.User{
		Name:  "new Name",
		Email: "new@mail.com",
	}
	res := req.Put(body)

	as.Equal(http.StatusOK, res.Result().StatusCode)

	updated := &models.User{}
	err = as.DB.Find(updated, user.ID)
	if err != nil {
		as.T().Fatal("failed finding records")
	}

	as.Equal(body.Email, updated.Email)
	as.Equal(body.Name, updated.Name)
}

func (as *ActionSuite) Test_Users_Delete() {
	user := &models.User{
		Name:     "user",
		Email:    "email@mail.com",
		Password: "password",
	}
	err := as.DB.Create(user)
	if err != nil {
		as.T().Fatal("error creating user")
	}

	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	req := as.JSON(fmt.Sprintf("/users/%d", user.ID))
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	res := req.Delete()

	as.Equal(http.StatusNoContent, res.Result().StatusCode)

	count, err := as.DB.Where("id = ?", user.ID).Count(&models.User{})
	if err != nil {
		as.T().Fatal("failed to count records")
	}

	as.Equal(0, count)
}

func (as *ActionSuite) Test_Users_Delete_Own_Account() {
	token, err := Login(as)
	if err != nil {
		as.Fail("token generation failed")
	}

	req := as.JSON(fmt.Sprintf("/users/%d", UserAdmin.ID))
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	res := req.Delete()

	as.Equal(http.StatusBadRequest, res.Result().StatusCode)

	count, err := as.DB.Where("id = ?", UserAdmin.ID).Count(&models.User{})
	if err != nil {
		as.T().Fatal("failed to count records")
	}

	as.Equal(1, count)
}
