package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/fr"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
	fr_translations "gopkg.in/go-playground/validator.v9/translations/fr"
)

// User contains user information
type User struct {
	FirstName      string     `json:"first_name" validate:"required"`
	LastName       string     `json:"last_name" validate:"required"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `validate:"required,email"`
	FavouriteColor string     `validate:"hexcolor|rgb|rgba"`
	Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
}

// Address houses a users address information
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

// use a single instance , it caches struct info
var (
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

func main() {

	en := en.New()
	fr := fr.New()
	uni = ut.New(en, en, fr)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	transEn, _ := uni.GetTranslator("en")
	transFr, _ := uni.GetTranslator("fr")
	transEn.Add("{{first_name}}", "First Name", false)
	transFr.Add("{{first_name}}", "Prénom", false)
	transEn.Add("{{last_name}}", "Last Name", false)
	transFr.Add("{{last_name}}", "Nom de famille", false)
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return "{{" + name + "}}"
	})

	en_translations.RegisterDefaultTranslations(validate, transEn)
	fr_translations.RegisterDefaultTranslations(validate, transFr)

	// build 'User' info, normally posted data etc...
	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		Phone:  "none",
		City:   "Unknown",
	}

	user := &User{
		FirstName:      "",
		LastName:       "",
		Age:            45,
		Email:          "",
		FavouriteColor: "#000",
		Addresses:      []*Address{address},
	}

	// returns InvalidValidationError for bad validation input, nil or ValidationErrors ( []FieldError )
	err := validate.Struct(user)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return
		}

		errMsg := make(map[string]string)

		for _, err := range err.(validator.ValidationErrors) {

			/*
				fmt.Println("Namespace: " + err.Namespace())
				fmt.Println("Field: " + err.Field())
				fmt.Println("StructNamespace: " + err.StructNamespace()) // can differ when a custom TagNameFunc is registered or
				fmt.Println("StructField: " + err.StructField())         // by passing alt name to ReportError like below
				fmt.Println("Tag: " + err.Tag())
				fmt.Println("ActualTag: " + err.ActualTag())
				fmt.Println("Kind: ", err.Kind())
				fmt.Println("Type: ", err.Type())
				fmt.Println("Value: ", err.Value())
				fmt.Println("Param: " + err.Param())
				fmt.Println(err.Translate(transFr))
				fmt.Println()
			*/
			jsonKey := err.Field()
			fieldName, _ := transFr.T(jsonKey)
			message := strings.Replace(err.Translate(transFr), jsonKey, fieldName, -1)
			jsonKey = jsonKey[2 : len(jsonKey)-2]
			errMsg[jsonKey] = message
			fmt.Println(jsonKey, ":", errMsg[jsonKey])
		}

		// from here you can create your own error messages in whatever language you wish
		return
	}

	// save user to database
}
