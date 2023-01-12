package rules

import (
	"coke/models"
	"fmt"
	"log"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Unique struct {
	Name   string
	Field  string
	Except int
	Model  interface{}
}

func (v *Unique) IsValid(errors *validate.Errors) {
	query := models.DB.Where(fmt.Sprintf("%s = ?", v.Name), v.Field)

	if v.Except != 0 {
		query = query.Where("id != ?", v.Except)
	}

	count, err := query.Count(v.Model)
	if err != nil {
		errors.Add(validators.GenerateKey(v.Name), "Could not get records on database")
		log.Println(err)
		return
	}

	if count > 0 {
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The %s has already been taken", v.Name))
	}
}
