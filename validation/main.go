package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gitlab.com/silenteer/titan/validation/v"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func main() {
	err := v.ValidateStruct1()

	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return
		}

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println("Namespace = ", err.Namespace())
			fmt.Println("Field = ", err.Field())
			fmt.Println("StructNamespace = ", err.StructNamespace())
			fmt.Println("StructField = ", err.StructField())
			fmt.Println("Tag = ", err.Tag())
			fmt.Println("ActualTag =", err.ActualTag())
			fmt.Println("Kind = ", err.Kind())
			fmt.Println("Type =", err.Type())
			fmt.Println("Value = ", err.Value())
			fmt.Println("Param =", err.Param())
			fmt.Println("------------------------------------------------------")
		}

		// from here you can create your own error messages in whatever language you wish
		return
	}
}
