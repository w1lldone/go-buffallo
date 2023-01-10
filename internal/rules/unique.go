package rules

import (
	"coke/models"
	"fmt"
	"log"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Unique struct {
	Name  string
	Field string
	Table string
	Model interface{}
}

func (v *Unique) IsValid(errors *validate.Errors) {
	count, err := models.DB.Where(fmt.Sprintf("%s = ?", v.Name), v.Field).Count(v.Model)
	if err != nil {
		errors.Add(validators.GenerateKey(v.Name), "Could not get records on database")
		log.Println(err)
		return
	}

	if count > 0 {
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The %s has already been taken", v.Name))
	}
}
