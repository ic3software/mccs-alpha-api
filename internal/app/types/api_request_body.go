package types

import (
	"errors"
	"strconv"
	"strings"

	"unicode"

	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SignupReqBody struct {
	Email                 string `json:"email"`
	Password              string `json:"password"`
	FirstName             string `json:"firstName"`
	LastName              string `json:"lastName"`
	UserPhone             string `json:"userPhone"`
	ShowRecentMatchedTags *bool  `json:"showTagsMatchedSinceLastLogin"`
	DailyNotification     *bool  `json:"dailyEmailMatchNotification"`
	EntityName            string `json:"entityName"`
	IncType               string `json:"incType"`
	CompanyNumber         string `json:"companyNumber"`
	EntityPhone           string `json:"entityPhone"`
	Website               string `json:"website"`
	Turnover              int    `json:"turnover"`
	Description           string `json:"description"`
	LocationAddress       string `json:"locationAddress"`
	LocationCity          string `json:"locationCity"`
	LocationRegion        string `json:"locationRegion"`
	LocationPostalCode    string `json:"locationPostalCode"`
	LocationCountry       string `json:"locationCountry"`
}

func (req *SignupReqBody) Validate() []error {
	errs := []error{}

	errs = append(errs, validateEmail(req.Email)...)
	errs = append(errs, validatePassword(req.Password)...)

	user := User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	}
	errs = append(errs, user.Validate()...)

	return errs
}

type LoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginReqBody) Validate() []error {
	errs := []error{}
	if req.Password == "" {
		errs = append(errs, errors.New("Password is missing."))
	}
	return errs
}

type ResetPasswordReqBody struct {
	Password string `json:"password"`
}

func (req *ResetPasswordReqBody) Validate() []error {
	errs := []error{}
	errs = append(errs, validatePassword(req.Password)...)
	return errs
}

type PasswordChange struct {
	Password string `json:"password"`
}

func (req *PasswordChange) Validate() []error {
	errs := []error{}
	errs = append(errs, validatePassword(req.Password)...)
	return errs
}

type UpdateUserReqBody struct {
	ID                            string `json:"id"`
	Email                         string `json:"email"`
	FirstName                     string `json:"firstName"`
	LastName                      string `json:"lastName"`
	UserPhone                     string `json:"userPhone"`
	DailyEmailMatchNotification   *bool  `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin *bool  `json:"showTagsMatchedSinceLastLogin"`
}

func (req *UpdateUserReqBody) Validate() []error {
	errs := []error{}

	if req.ID != "" {
		errs = append(errs, errors.New("Your ID cannot be changed."))
	}
	if req.Email != "" {
		errs = append(errs, errors.New("Your email address can only be changed by an administrator."))
	}

	user := User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	}
	errs = append(errs, user.Validate()...)

	return errs
}

type UpdateUserEntityReqBody struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	EntityName         string   `json:"entityName"`
	EntityPhone        string   `json:"entityPhone"`
	IncType            string   `json:"incType"`
	CompanyNumber      string   `json:"companyNumber"`
	Website            string   `json:"website"`
	Turnover           int      `json:"turnover"`
	Description        string   `json:"description"`
	LocationAddress    string   `json:"locationAddress"`
	LocationCity       string   `json:"locationCity"`
	LocationRegion     string   `json:"locationRegion"`
	LocationPostalCode string   `json:"locationPostalCode"`
	LocationCountry    string   `json:"locationCountry"`
	Offers             []string `json:"offers"`
	Wants              []string `json:"wants"`
}

func (req *UpdateUserEntityReqBody) Validate() []error {
	errs := []error{}

	if req.ID != "" {
		errs = append(errs, errors.New("The entity ID cannot be changed."))
	}
	if req.Status != "" {
		errs = append(errs, errors.New("The status cannot be changed."))
	}

	entity := Entity{
		EntityName:         req.EntityName,
		EntityPhone:        req.EntityPhone,
		IncType:            req.IncType,
		CompanyNumber:      req.CompanyNumber,
		Website:            req.Website,
		Turnover:           req.Turnover,
		Description:        req.Description,
		LocationCity:       req.LocationCity,
		LocationCountry:    req.LocationCountry,
		LocationAddress:    req.LocationAddress,
		LocationRegion:     req.LocationRegion,
		LocationPostalCode: req.LocationPostalCode,
	}
	errs = append(errs, entity.Validate()...)
	errs = append(errs, validateTags(req.Offers)...)
	errs = append(errs, validateTags(req.Wants)...)

	return errs
}

type AddToFavoriteReqBody struct {
	AddToEntityID    string `json:"addToEntityID"`
	FavoriteEntityID string `json:"favoriteEntityID"`
	Favorite         *bool  `json:"favorite"`
}

func (req *AddToFavoriteReqBody) Validate() []error {
	errs := []error{}
	_, err := primitive.ObjectIDFromHex(req.AddToEntityID)
	if err != nil {
		errs = append(errs, errors.New("AddToEntityID is wrong"))
	}
	_, err = primitive.ObjectIDFromHex(req.FavoriteEntityID)
	if err != nil {
		errs = append(errs, errors.New("FavoriteEntityID is wrong"))
	}
	if req.Favorite == nil {
		errs = append(errs, errors.New("Favorite is nil"))
	}
	return errs
}

func (req *AddToFavoriteReqBody) GetIsFavorite() bool {
	return *req.Favorite
}

func (req *AddToFavoriteReqBody) GetAddToEntityID() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(req.AddToEntityID)
	return id
}

func (req *AddToFavoriteReqBody) GetFavoriteEntityID() primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(req.FavoriteEntityID)
	return id
}

func validateTags(tags []string) []error {
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

func validateEmail(email string) []error {
	errs := []error{}
	email = strings.ToLower(email)
	emailMaxLen := viper.GetInt("validate.email.maxLen")
	if email == "" {
		errs = append(errs, errors.New("Email is missing."))
	} else if len(email) > emailMaxLen {
		errs = append(errs, errors.New("Email address length cannot exceed "+strconv.Itoa(emailMaxLen)+" characters."))
	} else if util.IsInValidEmail(email) {
		errs = append(errs, errors.New("Email is invalid."))
	}
	return errs
}

func validatePassword(password string) []error {
	minLen, hasLetter, hasNumber, hasSpecial := viper.GetInt("validate.password.minLen"), false, false, false

	errs := []error{}

	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsNumber(ch):
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if password == "" {
		errs = append(errs, errors.New("Password is missing."))
	} else if len(password) < minLen {
		errs = append(errs, errors.New("Password must be at least "+strconv.Itoa(minLen)+" characters long."))
	} else if len(password) > 100 {
		errs = append(errs, errors.New("Password cannot exceed 100 characters."))
	}
	if !hasLetter {
		errs = append(errs, errors.New("Password must have at least one letter."))
	}
	if !hasNumber {
		errs = append(errs, errors.New("Password must have at least one numeric value."))
	}
	if !hasSpecial {
		errs = append(errs, errors.New("Password must have at least one special character."))
	}

	return errs
}