package validate

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ic3network/mccs-alpha-api/internal/app/types"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/util"
	"github.com/spf13/viper"
)

var (
	emailMaxLen = viper.GetInt("validate.email.maxLen")
)

func SignUp(req *types.SignupReqBody) []error {
	errs := []error{}

	errs = append(errs, checkEmail(req.Email)...)
	errs = append(errs, validatePassword(req.Password)...)
	errs = append(errs, checkUser(&types.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	})...)

	return errs
}

func Login(password string) []error {
	errs := []error{}
	if password == "" {
		errs = append(errs, errors.New("Password is missing."))
	}
	return errs
}

func ResetPassword(password string) []error {
	errs := []error{}
	errs = append(errs, validatePassword(password)...)
	return errs
}

func UpdateUser(update *types.UpdateUserReqBody) []error {
	errs := []error{}

	if update.ID != "" {
		errs = append(errs, errors.New("Your ID cannot be changed."))
	}
	if update.Email != "" {
		errs = append(errs, errors.New("Your email address can only be changed by an administrator."))
	}
	errs = append(errs, checkUser(&types.User{
		FirstName: update.FirstName,
		LastName:  update.LastName,
		Telephone: update.UserPhone,
	})...)

	return errs
}

func UpdateUserEntity(update *types.UpdateUserEntityReqBody) []error {
	errs := []error{}

	if update.ID != "" {
		errs = append(errs, errors.New("The entity ID cannot be changed."))
	}
	if update.Status != "" {
		errs = append(errs, errors.New("The status cannot be changed."))
	}
	errs = append(errs, checkEntity(&types.Entity{
		EntityName:         update.EntityName,
		EntityPhone:        update.EntityPhone,
		IncType:            update.IncType,
		CompanyNumber:      update.CompanyNumber,
		Website:            update.Website,
		Turnover:           update.Turnover,
		Description:        update.Description,
		LocationCity:       update.LocationCity,
		LocationCountry:    update.LocationCountry,
		LocationAddress:    update.LocationAddress,
		LocationRegion:     update.LocationRegion,
		LocationPostalCode: update.LocationPostalCode,
	})...)
	errs = append(errs, checkTags(update.Offers)...)
	errs = append(errs, checkTags(update.Wants)...)

	return errs
}

func SearchEntity(query *types.SearchEntityQuery) []error {
	errs := []error{}
	return errs
}

func SearchTag(query *types.SearchTagQuery) []error {
	errs := []error{}
	return errs
}

func checkEmail(email string) []error {
	errs := []error{}
	email = strings.ToLower(email)
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if util.IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}
	return errs
}

func checkUser(user *types.User) []error {
	errs := []error{}
	if len(user.FirstName) > 100 {
		errs = append(errs, errors.New("First name length cannot exceed 100 characters."))
	}
	if len(user.LastName) > 100 {
		errs = append(errs, errors.New("Last name length cannot exceed 100 characters."))
	}
	if len(user.Telephone) > 25 {
		errs = append(errs, errors.New("Telephone length cannot exceed 25 characters."))
	}
	return errs
}

func checkEntity(entity *types.Entity) []error {
	errs := []error{}
	if len(entity.EntityName) > 100 {
		errs = append(errs, errors.New("Entity name length cannot exceed 100 characters."))
	}
	if len(entity.EntityPhone) > 25 {
		errs = append(errs, errors.New("Telephone length cannot exceed 25 characters."))
	}
	if len(entity.IncType) > 25 {
		errs = append(errs, errors.New("Incorporation type length cannot exceed 25 characters."))
	}
	if len(entity.CompanyNumber) > 20 {
		errs = append(errs, errors.New("Company number length cannot exceed 20 characters."))
	}
	if len(entity.Website) > 100 {
		errs = append(errs, errors.New("Website URL length cannot exceed 100 characters."))
	}
	if len(entity.Description) > 500 {
		errs = append(errs, errors.New("Description length cannot exceed 500 characters."))
	}
	if len(entity.LocationCountry) > 10 {
		errs = append(errs, errors.New("Country length cannot exceed 50 characters."))
	}
	if len(entity.LocationCity) > 10 {
		errs = append(errs, errors.New("City length cannot exceed 50 characters."))
	}
	if len(entity.LocationAddress) > 255 {
		errs = append(errs, errors.New("Address length cannot exceed 255 characters."))
	}
	if len(entity.LocationRegion) > 50 {
		errs = append(errs, errors.New("Region length cannot exceed 50 characters."))
	}
	if len(entity.LocationPostalCode) > 10 {
		errs = append(errs, errors.New("Postal code length cannot exceed 10 characters."))
	}
	return errs
}

func checkTags(tags []string) []error {
	errs := []error{}
	if len(tags) > 10 {
		errs = append(errs, errors.New("You can only specify a maximum of 10 tags."))
	}
	for _, tag := range tags {
		if len(tag) > 50 {
			errs = append(errs, errors.New("Tag length cannot exceed 50 characters."))
			break
		}
	}
	return errs
}
