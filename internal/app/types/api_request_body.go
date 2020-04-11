package types

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"unicode"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/internal/pkg/bcrypt"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewSignupReqBody(r *http.Request) (*SignupReqBody, error) {
	var req SignupReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

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

func NewLoginReqBody(r *http.Request) (*LoginReqBody, []error) {
	var req LoginReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.Validate()
}

type LoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginReqBody) Validate() []error {
	errs := []error{}
	if req.Email == "" {
		errs = append(errs, errors.New("Please specify an email address."))
	}
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

func NewUpdateUserReqBody(r *http.Request) (*UpdateUserReqBody, []error) {
	var req UpdateUserReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.Validate()
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

func NewUpdateUserEntityReqBody(r *http.Request) (*UpdateUserEntityReqBody, error) {
	var req UpdateUserEntityReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	req.Offers, req.Wants = util.FormatTags(req.Offers), util.FormatTags(req.Wants)
	return &req, nil
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

func NewAddToFavoriteReqBody(r *http.Request) (*AddToFavoriteReqBody, error) {
	var req AddToFavoriteReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
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
		errs = append(errs, errors.New("add_to_entity_id is incorrect."))
	}
	_, err = primitive.ObjectIDFromHex(req.FavoriteEntityID)
	if err != nil {
		errs = append(errs, errors.New("favorite_entity_id is incorrect."))
	}
	if req.Favorite == nil {
		errs = append(errs, errors.New("Favorite must be specified."))
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

func NewEmailReqBody(r *http.Request) (*EmailReqBody, error) {
	var req EmailReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
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
		errs = append(errs, errors.New("sender_entity_id is incorrect."))
	}
	_, err = primitive.ObjectIDFromHex(req.ReceiverEntityID)
	if err != nil {
		errs = append(errs, errors.New("receiver_entity_id is incorrect."))
	}
	if len(req.Body) == 0 {
		errs = append(errs, errors.New("Email body is empty."))
	}

	return errs
}

type TransferReqBody struct {
	// User Inputs
	TransferType           string
	InitiatorAccountNumber string
	ReceiverAccountNumber  string
	Amount                 float64
	Description            string

	InitiatorEmail      string
	InitiatorEntityName string

	ReceiverEmail      string
	ReceiverEntityName string

	FromAccountNumber string
	FromEmail         string
	FromEntityName    string
	FromStatus        string

	ToAccountNumber string
	ToEmail         string
	ToEntityName    string
	ToStatus        string

	InitiatorEntity *Entity
	ReceiverEntity  *Entity
}

func (req *TransferReqBody) Validate() []error {
	errs := []error{}

	if req.TransferType != constant.TransferType.In && req.TransferType != constant.TransferType.Out {
		errs = append(errs, errors.New("Transfer can be only 'in' or 'out'."))
	}

	if req.InitiatorAccountNumber == "" {
		errs = append(errs, errors.New("Initiator is empty."))
	} else {
		err := goluhn.Validate(req.InitiatorAccountNumber)
		if err != nil {
			errs = append(errs, errors.New("Initiator account number is invalid."))
		}
	}

	if req.ReceiverAccountNumber == "" {
		errs = append(errs, errors.New("Receiver is empty."))
	} else {
		err := goluhn.Validate(req.ReceiverAccountNumber)
		if err != nil {
			errs = append(errs, errors.New("Receiver account number is wrong."))
		}
	}

	// Amount should be positive value and with up to two decimal places.
	if req.Amount <= 0 || !util.IsDecimalValid(req.Amount) {
		errs = append(errs, errors.New("Please enter a valid numeric amount to send with up to two decimal places."))
	}

	// Only allow transfers with accounts that also have "trading-accepted" status
	if req.FromStatus != constant.Trading.Accepted {
		errs = append(errs, errors.New("Sender is not a trading member. Transfers can only be made when both entities have trading member status."))
	} else if req.ToStatus != constant.Trading.Accepted {
		errs = append(errs, errors.New("Recipient is not a trading member. Transfers can only be made when both entities have trading member status."))
	}

	// Check if the user is doing the transaction to himself.
	if req.FromAccountNumber == req.ToAccountNumber {
		errs = append(errs, errors.New("You cannot create a transaction with yourself."))
	}

	return errs
}

type UpdateTransferReqBody struct {
	TransferID string
	Action     string
	Reason     string

	LoggedInUserID string

	Journal        *Journal
	InitiateEntity *Entity
	FromEntity     *Entity
	ToEntity       *Entity
}

func (req *UpdateTransferReqBody) Validate() []error {
	errs := []error{}

	if req.Action != "accept" && req.Action != "reject" && req.Action != "cancel" {
		errs = append(errs, errors.New("Please enter a valid action."))
	}
	if req.Journal.Status == constant.Transfer.Completed {
		errs = append(errs, errors.New("The transaction has already been completed by the counterparty."))
	} else if req.Journal.Status == constant.Transfer.Cancelled {
		errs = append(errs, errors.New("The transaction has already been cancelled by the counterparty."))
	}

	return errs
}

func NewSearchEntityReqBody(q url.Values) (*SearchEntityReqBody, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchEntityReqBody{
		QueryingEntityID: q.Get("querying_entity_id"),
		Page:             page,
		PageSize:         pageSize,
		EntityName:       q.Get("entity_name"),
		Category:         q.Get("category"),
		Offers:           util.ToSearchTags(q.Get("offers")),
		Wants:            util.ToSearchTags(q.Get("wants")),
		TaggedSince:      util.ParseTime(q.Get("tagged_since")),
		FavoritesOnly:    q.Get("favorites_only") == "true",
		Statuses: []string{
			constant.Entity.Accepted,
			constant.Trading.Pending,
			constant.Trading.Accepted,
			constant.Trading.Rejected,
		},
	}, nil
}

type SearchEntityReqBody struct {
	QueryingEntityID string
	Page             int
	PageSize         int
	EntityName       string
	Wants            []string
	Offers           []string
	Category         string
	FavoriteEntities []primitive.ObjectID
	FavoritesOnly    bool
	TaggedSince      time.Time
	Statuses         []string // accepted", "pending", rejected", "tradingPending", "tradingAccepted", "tradingRejected"

	LocationCountry string
	LocationCity    string
}

func (query *SearchEntityReqBody) Validate() []error {
	errs := []error{}

	if query.FavoritesOnly == true && query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}

	if !query.TaggedSince.IsZero() && len(query.Wants) == 0 && len(query.Offers) == 0 {
		errs = append(errs, errors.New("Please specify an offer or want tag."))
	}

	return errs
}

func NewSearchTagReqBody(q url.Values) (*SearchTagReqBody, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchTagReqBody{
		Fragment: q.Get("fragment"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

type SearchTagReqBody struct {
	Fragment string `json:"fragment"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (q *SearchTagReqBody) Validate() []error {
	errs := []error{}
	return errs
}

func NewSearchCategoryReqBody(q url.Values) (*SearchCategoryReqBody, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchCategoryReqBody{
		Fragment: q.Get("fragment"),
		Prefix:   q.Get("prefix"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

type SearchCategoryReqBody struct {
	Fragment string `json:"fragment"`
	Prefix   string `json:"prefix"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (query *SearchCategoryReqBody) Validate() []error {
	errs := []error{}
	return errs
}

type SearchTransferReqBody struct {
	Page             int
	PageSize         int
	Status           string
	QueryingEntityID string

	QueryingAccountNumber string
	Offset                int
}

func (query *SearchTransferReqBody) Validate() []error {
	errs := []error{}

	if query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}
	if query.Status != "all" && query.Status != "initiated" && query.Status != "completed" && query.Status != "cancelled" {
		errs = append(errs, errors.New("Please specify valid status."))
	}

	return errs
}

type BalanceReqBody struct {
	QueryingEntityID string
}

func (query *BalanceReqBody) Validate() []error {
	errs := []error{}

	if query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}

	return errs
}

// Admin

func NewAdminUpdateEntityReqBody(r *http.Request) (*AdminUpdateEntityReqBody, error) {
	var req AdminUpdateEntityReqBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, err
	}

	vars := mux.Vars(r)
	entityID, err := primitive.ObjectIDFromHex(vars["entityID"])
	if err != nil {
		return nil, err
	}
	req.EntityID = entityID
	req.Offers, req.Wants, req.Categories = util.FormatTags(req.Offers), util.FormatTags(req.Wants), util.FormatTags(req.Categories)

	return &req, nil
}

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

type AdminUpdateCategoryReqBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAdminUpdateCategoryReqBody(r *http.Request) (*AdminUpdateCategoryReqBody, []error) {
	var req AdminUpdateCategoryReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	req.ID = mux.Vars(r)["id"]
	return &req, req.validate()
}

func (req *AdminUpdateCategoryReqBody) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the category name."))
	}
	return errs
}

type AdminCreateCategoryReqBody struct {
	Name string `json:"name"`
}

func NewAdminCreateCategoryReqBody(r *http.Request) (*AdminCreateCategoryReqBody, []error) {
	var req AdminCreateCategoryReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	return &req, req.validate()
}

func (req *AdminCreateCategoryReqBody) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the category name."))
	}
	return errs
}

type AdminDeleteCategoryReqBody struct {
	ID primitive.ObjectID `json:"name"`
}

func NewAdminDeleteCategoryReqBody(r *http.Request) (*AdminDeleteCategoryReqBody, []error) {
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, []error{errors.New("Please enter category id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, []error{errors.New("Please enter valid category id.")}
	}
	return &AdminDeleteCategoryReqBody{
		ID: objectID,
	}, nil
}

type AdminCreateTagReqBody struct {
	Name string `json:"name"`
}

func NewAdminCreateTagReqBody(r *http.Request) (*AdminCreateTagReqBody, []error) {
	var req AdminCreateTagReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	return &req, req.validate()
}

func (req *AdminCreateTagReqBody) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the tag name."))
	}
	return errs
}

type AdminUpdateTagReqBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAdminUpdateTagReqBody(r *http.Request) (*AdminUpdateTagReqBody, []error) {
	var req AdminUpdateTagReqBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	req.ID = mux.Vars(r)["id"]
	return &req, req.validate()
}

func (req *AdminUpdateTagReqBody) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the tag name."))
	}
	return errs
}

type AdminDeleteTagReqBody struct {
	ID primitive.ObjectID `json:"name"`
}

func NewAdminDeleteTagReqBody(r *http.Request) (*AdminDeleteTagReqBody, []error) {
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, []error{errors.New("Please enter tag id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, []error{errors.New("Please enter valid tag id.")}
	}
	return &AdminDeleteTagReqBody{
		ID: objectID,
	}, nil
}

type AdminGetUser struct {
	UserID primitive.ObjectID
}

func NewAdminGetUserReqBody(r *http.Request) (*AdminGetUser, []error) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		return nil, []error{errors.New("Please enter user id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, []error{errors.New("Please enter valid user id.")}
	}
	return &AdminGetUser{
		UserID: objectID,
	}, nil
}

type AdminUpdateUser struct {
	UserID                        primitive.ObjectID
	Email                         string
	FirstName                     string
	LastName                      string
	UserPhone                     string
	Password                      string
	DailyEmailMatchNotification   *bool
	ShowTagsMatchedSinceLastLogin *bool
}

type adminUpdateUser struct {
	Email                         string `json:"email"`
	FirstName                     string `json:"firstName"`
	LastName                      string `json:"lastName"`
	UserPhone                     string `json:"userPhone"`
	Password                      string `json:"password"`
	DailyEmailMatchNotification   *bool  `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin *bool  `json:"showTagsMatchedSinceLastLogin"`
}

func NewAdminUpdateUserReqBody(r *http.Request) (*AdminUpdateUser, []error) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		return nil, []error{errors.New("Please enter user id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, []error{errors.New("Please enter valid user id.")}
	}

	req := adminUpdateUser{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}

	errs := req.validate()
	if len(errs) != 0 {
		return nil, errs
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.Hash(req.Password)
		if err != nil {
			return nil, []error{err}
		}
		req.Password = hashedPassword
	}

	return &AdminUpdateUser{
		UserID:                        objectID,
		Email:                         req.Email,
		FirstName:                     req.FirstName,
		LastName:                      req.LastName,
		UserPhone:                     req.UserPhone,
		Password:                      req.Password,
		DailyEmailMatchNotification:   req.DailyEmailMatchNotification,
		ShowTagsMatchedSinceLastLogin: req.ShowTagsMatchedSinceLastLogin,
	}, nil
}

func (req *adminUpdateUser) validate() []error {
	errs := []error{}

	if req.Password != "" {
		errs = append(errs, validatePassword(req.Password)...)
	}

	user := User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	}
	errs = append(errs, user.Validate()...)

	return errs
}

type AdminDeleteUser struct {
	UserID primitive.ObjectID
}

func NewAdminDeleteUserReqBody(r *http.Request) (*AdminDeleteUser, []error) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		return nil, []error{errors.New("Please enter user id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, []error{errors.New("Please enter valid user id.")}
	}

	return &AdminDeleteUser{
		UserID: objectID,
	}, nil
}
