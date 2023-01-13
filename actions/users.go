package actions

import (
	"coke/internal/rules"
	"coke/models"
	"errors"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type UserResource struct{}

// UserIndex default implementation.
func (u UserResource) Index(c buffalo.Context) error {
	users := &models.Users{}
	query := models.DB.PaginateFromParams(c.Params())
	err := query.All(users)
	if err != nil {
		return err
	}

	response := Response{
		Data:   users,
		Status: "ok",
		Meta:   query.Paginator,
	}
	return c.Render(http.StatusOK, r.JSON(response))
}

func (u UserResource) Show(c buffalo.Context) error {
	userId := c.Param("user_id")
	user := models.User{}
	err := models.DB.Find(&user, userId)
	if err != nil {
		return err
	}

	response := Response{
		Data:   user,
		Status: "ok",
	}
	return c.Render(http.StatusOK, r.JSON(response))
}

type userJson struct {
	Name                 string    `json:"name"`
	Email                string    `json:"email"`
	Password             string    `json:"password"`
	PasswordConfirmation string    `json:"password_confirmation"`
	AccessLevel          nulls.Int `json:"access_level"`
}

func (u UserResource) Store(c buffalo.Context) error {
	userJson := &userJson{}
	if err := c.Bind(userJson); err != nil {
		return err
	}

	verr := validate.Validate(
		&validators.StringIsPresent{Field: userJson.Name, Name: "name"},
		&validators.StringLengthInRange{Name: "name", Field: userJson.Name, Min: 3, Max: 100},

		&validators.EmailIsPresent{Field: userJson.Email, Name: "email"},
		&rules.Unique{Name: "email", Field: userJson.Email, Model: &models.User{}},

		&validators.IntIsPresent{Field: userJson.AccessLevel.Int, Name: "access_level"},
		&validators.IntIsLessThan{Name: "access_level", Field: userJson.AccessLevel.Int, Compared: 5},

		&validators.StringIsPresent{Field: userJson.Password, Name: "password"},
		&validators.StringsMatch{Field: userJson.Password, Field2: userJson.PasswordConfirmation, Name: "password", Message: "Password and confirmation did not match."},
	)

	if verr.HasAny() {
		response := Response{
			Errors: verr.Errors,
			Status: "error",
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(response))
	}

	user := &models.User{
		Name:                 userJson.Name,
		Email:                userJson.Email,
		Password:             userJson.Password,
		PasswordConfirmation: userJson.PasswordConfirmation,
		AccessLevel:          userJson.AccessLevel,
	}
	err := models.DB.Create(user)
	if err != nil {
		return err
	}

	return c.Render(http.StatusCreated, r.JSON(user))
}

func (u UserResource) Update(c buffalo.Context) error {
	user := models.User{}
	err := models.DB.Find(&user, c.Param("user_id"))
	if err != nil {
		return err
	}

	form := &models.User{}
	if err := c.Bind(form); err != nil {
		return err
	}

	verr := validate.Validate(
		&validators.StringIsPresent{Field: form.Name, Name: "name"},
		&validators.StringIsPresent{Field: form.Email, Name: "email"},
		&validators.EmailIsPresent{Field: form.Email, Name: "email"},
		&validators.StringLengthInRange{Name: "name", Field: form.Name, Min: 3, Max: 100},
		&rules.Unique{Name: "email", Field: form.Email, Model: &models.User{}, Except: user.ID},
	)
	if verr.HasAny() {
		response := Response{
			Errors: verr.Errors,
			Status: "error",
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(response))
	}

	form.ID = user.ID
	err = models.DB.UpdateColumns(form, "name", "email")
	if err != nil {
		return err
	}

	response := Response{
		Data:   form,
		Status: "ok",
	}
	return c.Render(http.StatusOK, r.JSON(response))
}

func (u UserResource) Delete(c buffalo.Context) error {
	user := &models.User{}
	err := models.DB.Find(user, c.Param("user_id"))
	if err != nil {
		return err
	}

	auth := c.Value("auth").(*models.User)
	if auth.ID == user.ID {
		return c.Error(400, errors.New("can not delete your own account"))
	}

	err = models.DB.Destroy(user)
	if err != nil {
		return err
	}

	return c.Render(http.StatusNoContent, r.JSON(nil))
}
