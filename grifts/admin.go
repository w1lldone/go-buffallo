package grifts

import (
	"coke/internal/rules"
	"coke/models"
	"fmt"

	"github.com/gobuffalo/grift/grift"
	"github.com/gobuffalo/validate"
)

var _ = grift.Namespace("admin", func() {

	grift.Desc("seed", "Seed superadmin")
	grift.Add("seed", func(c *grift.Context) error {
		user := &models.User{
			Name:     "admin",
			Email:    "admin@mail.com",
			Password: "password",
		}

		verr := validate.Validate(
			&rules.Unique{Name: "email", Field: user.Email, Model: &models.User{}},
		)
		if verr.HasAny() {
			for _, err := range verr.Errors {
				fmt.Println(err)
			}

			return nil
		}

		err := models.DB.Create(user)
		if err != nil {
			fmt.Println("can not create an Admin")
			return nil
		}

		fmt.Println("Admin user has been created")

		return nil
	})

})
