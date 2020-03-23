package types

import (
	"errors"
	"strconv"

	"unicode"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/ic3network/mccs-alpha-api/global/constant"
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

	EntityName         string   `json:"entityName"`
	IncType            string   `json:"incType"`
	CompanyNumber      string   `json:"companyNumber"`
	EntityPhone        string   `json:"entityPhone"`
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

func (req *SignupReqBody) Validate() []error {
	errs := []error{}

	errs = append(errs, validateEmail(req.Email)...)
	errs = append(errs, validatePassword(req.Password)...)

	user := User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	}
	entity := Entity{
		Email:              req.Email,
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

	errs = append(errs, user.Validate()...)
	errs = append(errs, entity.Validate()...)
	errs = append(errs, validateTags(req.Offers)...)
	errs = append(errs, validateTags(req.Wants)...)

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
	EntityName         string   `json:"entityName"`
	Email              string   `json:"email"`
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
	// Not allow to change
	ID     string `json:"id"`
	Status string `json:"status"`
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
		Email:              req.Email,
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
	AddToEntityID    string `json:"add_to_entity_id"`
	FavoriteEntityID string `json:"favorite_entity_id"`
	Favorite         *bool  `json:"favorite"`
}

func (req *AddToFavoriteReqBody) Validate() []error {
	errs := []error{}

	_, err := primitive.ObjectIDFromHex(req.AddToEntityID)
	if err != nil {
		errs = append(errs, errors.New("add_to_entity_id is wrong"))
	}
	_, err = primitive.ObjectIDFromHex(req.FavoriteEntityID)
	if err != nil {
		errs = append(errs, errors.New("favorite_entity_id is wrong"))
	}
	if req.Favorite == nil {
		errs = append(errs, errors.New("favorite is nil"))
	}

	return errs
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
		errs = append(errs, errors.New("Password must have at least one number."))
	}
	if !hasSpecial {
		errs = append(errs, errors.New("Password must have at least one special character."))
	}

	return errs
}

type EmailReqBody struct {
	SenderEntityID   string `json:"sender_entity_id"`
	ReceiverEntityID string `json:"receiver_entity_id"`
	Body             string `json:"body"`
}

func (req *EmailReqBody) Validate() []error {
	errs := []error{}

	_, err := primitive.ObjectIDFromHex(req.SenderEntityID)
	if err != nil {
		errs = append(errs, errors.New("sender_entity_id is wrong"))
	}
	_, err = primitive.ObjectIDFromHex(req.ReceiverEntityID)
	if err != nil {
		errs = append(errs, errors.New("receiver_entity_id is wrong"))
	}
	if len(req.Body) == 0 {
		errs = append(errs, errors.New("body is empty"))
	}

	return errs
}

type TransferReqBody struct {
	Transfer               string  `json:"transfer"`
	InitiatorAccountNumber string  `json:"initiator"`
	ReceiverAccountNumber  string  `json:"receiver"`
	Amount                 float64 `json:"amount"`
	Description            string  `json:"description"`
}

func (req *TransferReqBody) Validate() []error {
	errs := []error{}

	if req.Transfer != constant.TransferType.In && req.Transfer != constant.TransferType.Out {
		errs = append(errs, errors.New("transfer can be only 'in' or 'out'"))
	}

	if req.InitiatorAccountNumber == "" {
		errs = append(errs, errors.New("initiator is empty"))
	} else {
		err := goluhn.Validate(req.InitiatorAccountNumber)
		if err != nil {
			errs = append(errs, errors.New("initiator account number is wrong"))
		}
	}

	if req.ReceiverAccountNumber == "" {
		errs = append(errs, errors.New("receiver is empty"))
	} else {
		err := goluhn.Validate(req.ReceiverAccountNumber)
		if err != nil {
			errs = append(errs, errors.New("receiver account number is wrong"))
		}
	}

	// Amount should be positive value and with up to two decimal places.
	if req.Amount <= 0 || !util.IsDecimalValid(req.Amount) {
		errs = append(errs, errors.New("Please enter a valid numeric amount to send with up to two decimal places."))
	}

	return errs
}

type UpdateTransferReqBody struct {
	TransferID string
	Action     string
	Reason     string

	Journal *Journal
	User    *User
}

func (req *UpdateTransferReqBody) Validate() []error {
	errs := []error{}

	if req.Action != "accept" && req.Action != "reject" && req.Action != "cancel" {
		errs = append(errs, errors.New("Please enter a valid action"))
	}
	if req.Journal.Status == constant.Transfer.Completed {
		errs = append(errs, errors.New("The transaction has already been completed by the counterparty."))
	} else if req.Journal.Status == constant.Transfer.Cancelled {
		errs = append(errs, errors.New("The transaction has already been cancelled by the counterparty."))
	}

	return errs
}

// Admin

type AdminUpdateEntityReqBody struct {
	EntityID           primitive.ObjectID `json:"entityID"`
	ID                 string             `json:"id"`
	Status             string             `json:"status"`
	EntityName         string             `json:"entityName"`
	Email              string             `json:"email"`
	EntityPhone        string             `json:"entityPhone"`
	IncType            string             `json:"incType"`
	CompanyNumber      string             `json:"companyNumber"`
	Website            string             `json:"website"`
	Turnover           int                `json:"turnover"`
	Description        string             `json:"description"`
	LocationAddress    string             `json:"locationAddress"`
	LocationCity       string             `json:"locationCity"`
	LocationRegion     string             `json:"locationRegion"`
	LocationPostalCode string             `json:"locationPostalCode"`
	LocationCountry    string             `json:"locationCountry"`
	Offers             []string           `json:"offers"`
	Wants              []string           `json:"wants"`
	Categories         []string           `json:"categories"`
}

func (req *AdminUpdateEntityReqBody) Validate() []error {
	errs := []error{}

	if req.ID != "" {
		errs = append(errs, errors.New("The entity ID cannot be changed."))
	}

	entity := Entity{
		Email:              req.Email,
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
		Categories:         req.Categories,
		Status:             req.Status,
	}
	errs = append(errs, entity.Validate()...)
	errs = append(errs, validateTags(req.Offers)...)
	errs = append(errs, validateTags(req.Wants)...)

	return errs
}
