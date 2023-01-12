package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (as *ActionSuite) Test_Users_Index() {
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

	as.App.Logger.Debugf("unmarshaled ", data["data"])

	users := data["data"].([]interface{})
	as.Equal(1, len(users))
}
