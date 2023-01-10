package actions

import (
	"coke/internal/rules"
	"coke/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
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

func (u UserResource) Store(c buffalo.Context) error {
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		return err
	}

	verr := validate.Validate(
		&validators.StringIsPresent{Field: user.Name, Name: "name"},
		&validators.EmailIsPresent{Field: user.Email, Name: "email"},
		&rules.Unique{Name: "email", Field: user.Email, Model: &models.User{}},
		&validators.StringsMatch{Field: user.Password, Field2: user.PasswordConfirmation, Name: "password", Message: "Password and confirmation did not match."},
	)

	if verr.HasAny() {
		response := Response{
			Errors: verr.Errors,
			Status: "error",
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(response))
	}

	err := models.DB.Create(user)
	if err != nil {
		return err
	}

	return c.Render(http.StatusCreated, r.JSON(user))
}
