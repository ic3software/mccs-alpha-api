package types

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"unicode"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/gorilla/mux"
	"github.com/ic3network/mccs-alpha-api/global/constant"
	"github.com/ic3network/mccs-alpha-api/util"
	"github.com/ic3network/mccs-alpha-api/util/bcrypt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /signup

func NewSignupReq(r *http.Request) (*SignupReq, []error) {
	var req SignupReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.validate()
}

type SignupReq struct {
	UserEmail        string   `json:"userEmail"`
	EntityEmail      string   `json:"entityEmail"`
	Password         string   `json:"password"`
	FirstName        string   `json:"firstName"`
	LastName         string   `json:"lastName"`
	UserPhone        string   `json:"userPhone"`
	EntityName       string   `json:"entityName"`
	IncType          string   `json:"incType"`
	CompanyNumber    string   `json:"companyNumber"`
	EntityPhone      string   `json:"entityPhone"`
	Website          string   `json:"website"`
	DeclaredTurnover *int     `json:"declaredTurnover"`
	Description      string   `json:"description"`
	Address          string   `json:"address"`
	City             string   `json:"city"`
	Region           string   `json:"region"`
	PostalCode       string   `json:"postalCode"`
	Country          string   `json:"country"`
	Offers           []string `json:"offers"`
	Wants            []string `json:"wants"`
	// flags
	ShowTagsMatchedSinceLastLogin      *bool `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail *bool `json:"receiveDailyMatchNotificationEmail"`
}

func (req *SignupReq) validate() []error {
	errs := []error{}

	errs = append(errs, util.ValidateEmail(req.UserEmail)...)
	errs = append(errs, validatePassword(req.Password)...)

	user := User{
		Email:     req.UserEmail,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Telephone: req.UserPhone,
	}
	entity := Entity{
		Email:            req.EntityEmail,
		Name:             req.EntityName,
		Telephone:        req.EntityPhone,
		IncType:          req.IncType,
		CompanyNumber:    req.CompanyNumber,
		Website:          req.Website,
		DeclaredTurnover: req.DeclaredTurnover,
		Description:      req.Description,
		City:             req.City,
		Country:          req.Country,
		Address:          req.Address,
		Region:           req.Region,
		PostalCode:       req.PostalCode,
	}

	errs = append(errs, user.Validate()...)
	errs = append(errs, entity.Validate()...)
	errs = append(errs, validateTags(req.Offers)...)
	errs = append(errs, validateTags(req.Wants)...)

	return errs
}

// POST /login

func NewLoginReq(r *http.Request) (*LoginReq, []error) {
	var req LoginReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.validate()
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginReq) validate() []error {
	errs := []error{}
	if req.Email == "" {
		errs = append(errs, errors.New("Please specify an email address."))
	}
	if req.Password == "" {
		errs = append(errs, errors.New("Password is missing."))
	}
	return errs
}

type ResetPasswordReq struct {
	Password string `json:"password"`
}

func (req *ResetPasswordReq) Validate() []error {
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

// PATCH /user

func NewUpdateUserReq(r *http.Request) (*UpdateUserReq, []error) {
	var req UpdateUserReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.validate()
}

type UpdateUserReq struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserPhone string `json:"userPhone"`
}

func (req *UpdateUserReq) validate() []error {
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

func NewUpdateUserEntityReq(j UpdateUserEntityJSON, originEntity *Entity) (*UpdateUserEntityReq, []error) {
	errs := j.validate()
	if len(errs) != 0 {
		return nil, errs
	}

	addedOffers := []string{}
	removedOffers := []string{}
	if j.Offers != nil {
		addedOffers, removedOffers = util.StringDiff(util.FormatTags(*j.Offers), TagFieldToNames(originEntity.Offers))
	}
	addedWants := []string{}
	removedWants := []string{}
	if j.Wants != nil {
		addedWants, removedWants = util.StringDiff(util.FormatTags(*j.Wants), TagFieldToNames(originEntity.Wants))
	}

	req := UpdateUserEntityReq{
		OriginEntity:     originEntity,
		Name:             j.Name,
		Telephone:        j.Telephone,
		Email:            j.Email,
		IncType:          j.IncType,
		CompanyNumber:    j.CompanyNumber,
		Website:          j.Website,
		DeclaredTurnover: j.DeclaredTurnover,
		Description:      j.Description,
		// Tags
		Offers:        j.Offers,
		AddedOffers:   addedOffers,
		RemovedOffers: removedOffers,
		Wants:         j.Wants,
		AddedWants:    addedWants,
		RemovedWants:  removedWants,
		// Address
		Address:    j.Address,
		City:       j.City,
		Region:     j.Region,
		PostalCode: j.PostalCode,
		Country:    j.Country,
		// flags
		ShowTagsMatchedSinceLastLogin:      j.ShowTagsMatchedSinceLastLogin,
		ReceiveDailyMatchNotificationEmail: j.ReceiveDailyMatchNotificationEmail,
	}

	return &req, nil
}

type UpdateUserEntityReq struct {
	OriginEntity     *Entity
	Name             string
	Email            string
	Telephone        string
	IncType          string
	CompanyNumber    string
	Website          string
	DeclaredTurnover *int
	Description      string
	// Address
	Address    string
	City       string
	Region     string
	PostalCode string
	Country    string
	// Tags
	Offers        *[]string
	AddedOffers   []string
	RemovedOffers []string
	Wants         *[]string
	AddedWants    []string
	RemovedWants  []string
	// flags
	ShowTagsMatchedSinceLastLogin      *bool `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail *bool `json:"receiveDailyMatchNotificationEmail"`
}

type UpdateUserEntityJSON struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Telephone        string `json:"telephone"`
	IncType          string `json:"incType"`
	CompanyNumber    string `json:"companyNumber"`
	Website          string `json:"website"`
	DeclaredTurnover *int   `json:"declaredTurnover"`
	Description      string `json:"description"`
	Address          string `json:"address"`
	City             string `json:"city"`
	Region           string `json:"region"`
	PostalCode       string `json:"postalCode"`
	Country          string `json:"country"`
	// Tags
	Offers *[]string `json:"offers"`
	Wants  *[]string `json:"wants"`
	// flags
	ShowTagsMatchedSinceLastLogin      *bool `json:"showTagsMatchedSinceLastLogin"`
	ReceiveDailyMatchNotificationEmail *bool `json:"receiveDailyMatchNotificationEmail"`
	// Not allow to change
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (req *UpdateUserEntityJSON) validate() []error {
	errs := []error{}

	if req.ID != "" {
		errs = append(errs, errors.New("The entity ID cannot be changed."))
	}
	if req.Status != "" {
		errs = append(errs, errors.New("The status cannot be changed."))
	}

	entity := Entity{
		Email:            req.Email,
		Name:             req.Name,
		Telephone:        req.Telephone,
		IncType:          req.IncType,
		CompanyNumber:    req.CompanyNumber,
		Website:          req.Website,
		DeclaredTurnover: req.DeclaredTurnover,
		Description:      req.Description,
		City:             req.City,
		Country:          req.Country,
		Address:          req.Address,
		Region:           req.Region,
		PostalCode:       req.PostalCode,
	}
	errs = append(errs, entity.Validate()...)
	if req.Offers != nil {
		errs = append(errs, validateTags(*req.Offers)...)
	}
	if req.Wants != nil {
		errs = append(errs, validateTags(*req.Wants)...)
	}

	return errs
}

func NewAddToFavoriteReq(r *http.Request) (*AddToFavoriteReq, []error) {
	var req AddToFavoriteReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.validate()
}

type AddToFavoriteReq struct {
	AddToEntityID    string `json:"add_to_entity_id"`
	FavoriteEntityID string `json:"favorite_entity_id"`
	Favorite         *bool  `json:"favorite"`
}

func (req *AddToFavoriteReq) validate() []error {
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

func NewEmailReq(r *http.Request) (*EmailReq, []error) {
	var req EmailReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	return &req, req.validate()
}

type EmailReq struct {
	SenderEntityID   string `json:"sender_entity_id"`
	ReceiverEntityID string `json:"receiver_entity_id"`
	Body             string `json:"body"`
}

func (req *EmailReq) validate() []error {
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

// POST /transfers

func NewTransferReq(userReq *TransferUserReq, initiatorEntity *Entity, receiverEntity *Entity) (*TransferReq, []error) {
	req := &TransferReq{
		TransferDirection:      userReq.TransferDirection,
		TransferType:           constant.TransferType.Transfer,
		Amount:                 userReq.Amount,
		Description:            userReq.Description,
		InitiatorAccountNumber: initiatorEntity.AccountNumber,
		InitiatorEmail:         initiatorEntity.Email,
		InitiatorEntityName:    initiatorEntity.Name,
		ReceiverAccountNumber:  receiverEntity.AccountNumber,
		ReceiverEmail:          receiverEntity.Email,
		ReceiverEntityName:     receiverEntity.Name,
		InitiatorEntity:        initiatorEntity,
		ReceiverEntity:         receiverEntity,
	}

	if req.TransferDirection == constant.TransferDirection.Out {
		req.FromAccountNumber = initiatorEntity.AccountNumber
		req.FromEmail = initiatorEntity.Email
		req.FromEntityName = initiatorEntity.Name
		req.FromStatus = initiatorEntity.Status

		req.ToAccountNumber = receiverEntity.AccountNumber
		req.ToEmail = receiverEntity.Email
		req.ToEntityName = receiverEntity.Name
		req.ToStatus = receiverEntity.Status
	}

	if req.TransferDirection == constant.TransferDirection.In {
		req.FromAccountNumber = receiverEntity.AccountNumber
		req.FromEmail = receiverEntity.Email
		req.FromEntityName = receiverEntity.Name
		req.FromStatus = receiverEntity.Status

		req.ToAccountNumber = initiatorEntity.AccountNumber
		req.ToEmail = initiatorEntity.Email
		req.ToEntityName = initiatorEntity.Name
		req.ToStatus = initiatorEntity.Status
	}

	return req, req.Validate()
}

type TransferUserReq struct {
	TransferDirection      string  `json:"transfer"`
	InitiatorAccountNumber string  `json:"initiator"`
	ReceiverAccountNumber  string  `json:"receiver"`
	Amount                 float64 `json:"amount"`
	Description            string  `json:"description"`
}

type TransferReq struct {
	TransferDirection string // "in" or "out"
	TransferType      string // "Transfer" / "AdminTranser"

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

func (req *TransferReq) Validate() []error {
	errs := []error{}

	if req.TransferDirection != constant.TransferDirection.In && req.TransferDirection != constant.TransferDirection.Out {
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

// GET /transfers

func NewSearchTransferQuery(r *http.Request, entity *Entity) (*SearchTransferReq, []error) {
	q := r.URL.Query()
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}
	query := &SearchTransferReq{
		Page:                  page,
		PageSize:              pageSize,
		Status:                q.Get("status"),
		QueryingEntityID:      q.Get("querying_entity_id"),
		QueryingAccountNumber: entity.AccountNumber,
		Offset:                (page - 1) * pageSize,
	}

	return query, query.validate()
}

type SearchTransferReq struct {
	Page                  int
	PageSize              int
	Status                string
	QueryingEntityID      string
	QueryingAccountNumber string
	Offset                int
}

func (req *SearchTransferReq) validate() []error {
	errs := []error{}

	if req.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}
	if req.Status != "all" && req.Status != "initiated" && req.Status != "completed" && req.Status != "cancelled" {
		errs = append(errs, errors.New("Please specify valid status."))
	}

	return errs
}

// PATCH /transfers

func NewUpdateTransferReq(
	r *http.Request,
	journal *Journal,
	initiateEntity *Entity,
	fromEntity *Entity,
	toEntity *Entity,
) (*UpdateTransferReq, []error) {
	var body struct {
		Action             string `json:"action"`
		CancellationReason string `json:"cancellationReason"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		if err == io.EOF {
			return nil, []error{errors.New("Please provide valid inputs.")}
		}
		return nil, []error{err}
	}

	req := UpdateTransferReq{
		TransferID:         mux.Vars(r)["transferID"],
		LoggedInUserID:     r.Header.Get("userID"),
		Action:             body.Action,
		CancellationReason: body.CancellationReason,
		Journal:            journal,
		InitiateEntity:     initiateEntity,
		FromEntity:         fromEntity,
		ToEntity:           toEntity,
	}

	return &req, req.Validate()
}

type UpdateTransferReq struct {
	TransferID         string
	Action             string
	CancellationReason string

	LoggedInUserID string

	Journal        *Journal
	InitiateEntity *Entity
	FromEntity     *Entity
	ToEntity       *Entity
}

func (req *UpdateTransferReq) Validate() []error {
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

// GET /entities

func NewSearchEntityReq(q url.Values) (*SearchEntityReq, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchEntityReq{
		QueryingEntityID: q.Get("querying_entity_id"),
		Page:             page,
		PageSize:         pageSize,
		Name:             q.Get("name"),
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

type SearchEntityReq struct {
	QueryingEntityID string
	Page             int
	PageSize         int
	Name             string
	Wants            []string
	Offers           []string
	Category         string
	FavoriteEntities []primitive.ObjectID
	FavoritesOnly    bool
	TaggedSince      time.Time
	Statuses         []string

	Country string
	City    string
}

func (query *SearchEntityReq) Validate() []error {
	errs := []error{}

	if query.FavoritesOnly == true && query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}

	if !query.TaggedSince.IsZero() && len(query.Wants) == 0 && len(query.Offers) == 0 {
		errs = append(errs, errors.New("Please specify an offer or want tag."))
	}

	return errs
}

// GET /entities/{entityID}

func NewGetEntityReq(r *http.Request) (*GetEntity, []error) {
	req := &GetEntity{
		SearchEntityID:   mux.Vars(r)["searchEntityID"],
		QueryingEntityID: r.URL.Query().Get("querying_entity_id"),
	}
	return req, req.validate()
}

type GetEntity struct {
	SearchEntityID   string
	QueryingEntityID string
}

func (q *GetEntity) validate() []error {
	errs := []error{}
	return errs
}

func NewSearchTagReq(q url.Values) (*SearchTagReq, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchTagReq{
		Fragment: q.Get("fragment"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

type SearchTagReq struct {
	Fragment string `json:"fragment"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (q *SearchTagReq) Validate() []error {
	errs := []error{}
	return errs
}

func NewSearchCategoryReq(q url.Values) (*SearchCategoryReq, error) {
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, err
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, err
	}
	return &SearchCategoryReq{
		Fragment: q.Get("fragment"),
		Prefix:   q.Get("prefix"),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

type SearchCategoryReq struct {
	Fragment string `json:"fragment"`
	Prefix   string `json:"prefix"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

func (query *SearchCategoryReq) Validate() []error {
	errs := []error{}
	return errs
}

// GET /balance

func NewBalanceQuery(r *http.Request) (*BalanceReq, []error) {
	req := BalanceReq{
		QueryingEntityID: r.URL.Query().Get("querying_entity_id"),
	}
	return &req, req.Validate()
}

type BalanceReq struct {
	QueryingEntityID string
}

func (query *BalanceReq) Validate() []error {
	errs := []error{}

	if query.QueryingEntityID == "" {
		errs = append(errs, errors.New("Please specify the querying_entity_id."))
	}

	return errs
}

// Admin

type AdminUpdateCategoryReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAdminUpdateCategoryReq(r *http.Request) (*AdminUpdateCategoryReq, []error) {
	var req AdminUpdateCategoryReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	req.ID = mux.Vars(r)["id"]
	return &req, req.validate()
}

func (req *AdminUpdateCategoryReq) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the category name."))
	}
	return errs
}

type AdminCreateCategoryReq struct {
	Name string `json:"name"`
}

func NewAdminCreateCategoryReq(r *http.Request) (*AdminCreateCategoryReq, []error) {
	var req AdminCreateCategoryReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	return &req, req.validate()
}

func (req *AdminCreateCategoryReq) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the category name."))
	}
	return errs
}

type AdminDeleteCategoryReq struct {
	ID primitive.ObjectID `json:"name"`
}

func NewAdminDeleteCategoryReq(r *http.Request) (*AdminDeleteCategoryReq, []error) {
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, []error{errors.New("Please enter category id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, []error{errors.New("Please enter valid category id.")}
	}
	return &AdminDeleteCategoryReq{
		ID: objectID,
	}, nil
}

type AdminCreateTagReq struct {
	Name string `json:"name"`
}

func NewAdminCreateTagReq(r *http.Request) (*AdminCreateTagReq, []error) {
	var req AdminCreateTagReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	return &req, req.validate()
}

func (req *AdminCreateTagReq) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the tag name."))
	}
	return errs
}

type AdminUpdateTagReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAdminUpdateTagReq(r *http.Request) (*AdminUpdateTagReq, []error) {
	var req AdminUpdateTagReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return nil, []error{err}
	}
	req.Name = util.InputToTag(req.Name)
	req.ID = mux.Vars(r)["id"]
	return &req, req.validate()
}

func (req *AdminUpdateTagReq) validate() []error {
	errs := []error{}
	if req.Name == "" {
		errs = append(errs, errors.New("Please enter the tag name."))
	}
	return errs
}

type AdminDeleteTagReq struct {
	ID primitive.ObjectID `json:"name"`
}

func NewAdminDeleteTagReq(r *http.Request) (*AdminDeleteTagReq, []error) {
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, []error{errors.New("Please enter tag id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, []error{errors.New("Please enter valid tag id.")}
	}
	return &AdminDeleteTagReq{
		ID: objectID,
	}, nil
}

type AdminGetUser struct {
	UserID primitive.ObjectID
}

func NewAdminGetUserReq(r *http.Request) (*AdminGetUser, []error) {
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

// PATCH /admin/users/{userID}

func NewAdminUpdateUserReq(j AdminUpdateUserJSON, originUser *User) (*AdminUpdateUserReq, []error) {
	errs := j.validate()
	if len(errs) != 0 {
		return nil, errs
	}

	addedEntities := []string{}
	removedEntities := []string{}
	if j.Entities != nil {
		addedEntities, removedEntities = util.StringDiff(*j.Entities, util.ToIDStrings(originUser.Entities))
	}

	req := AdminUpdateUserReq{
		OriginUser:      originUser,
		Email:           j.Email,
		FirstName:       j.FirstName,
		LastName:        j.LastName,
		UserPhone:       j.UserPhone,
		Password:        j.Password,
		Entity:          j.Entities,
		AddedEntities:   util.ToObjectIDs(addedEntities),
		RemovedEntities: util.ToObjectIDs(removedEntities),
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.Hash(req.Password)
		if err != nil {
			return nil, []error{err}
		}
		req.Password = hashedPassword
	}

	return &req, nil
}

type AdminUpdateUserReq struct {
	OriginUser *User
	Email      string
	FirstName  string
	LastName   string
	UserPhone  string
	Password   string
	// Entity
	Entity          *[]string
	AddedEntities   []primitive.ObjectID
	RemovedEntities []primitive.ObjectID
}

type AdminUpdateUserJSON struct {
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	UserPhone string    `json:"userPhone"`
	Password  string    `json:"password"`
	Entities  *[]string `json:"entities"`
}

func (req *AdminUpdateUserJSON) validate() []error {
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

func NewAdminDeleteUserReq(r *http.Request) (*AdminDeleteUser, []error) {
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

type AdminSearchUserReq struct {
	Email    string `json:"email"`
	LastName string `json:"last_name"`
	Page     int
	PageSize int
}

func NewAdminSearchUserReq(r *http.Request) (*AdminSearchUserReq, []error) {
	q := r.URL.Query()

	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}

	req := AdminSearchUserReq{
		Email:    q.Get("email"),
		LastName: q.Get("last_name"),
		Page:     page,
		PageSize: pageSize,
	}

	return &req, req.validate()
}

func (req *AdminSearchUserReq) validate() []error {
	errs := []error{}
	return errs
}

// GET /admin/entities

func NewAdminSearchEntityReq(r *http.Request) (*AdminSearchEntityReq, []error) {
	q := r.URL.Query()

	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}
	balance, err := util.ToFloat64(q.Get("balance"))
	if err != nil {
		return nil, []error{err}
	}
	maxPosBal, err := util.ToFloat64(q.Get("max_pos_bal"))
	if err != nil {
		return nil, []error{err}
	}
	maxNegBal, err := util.ToFloat64(q.Get("max_neg_bal"))
	if err != nil {
		return nil, []error{err}
	}
	statuses, err := util.AdminMapEntityStatus(q.Get("status"))
	if err != nil {
		return nil, []error{err}
	}

	req := &AdminSearchEntityReq{
		Page:          page,
		PageSize:      pageSize,
		Name:          q.Get("name"),
		Email:         q.Get("email"),
		Statuses:      statuses,
		Offers:        util.ToSearchTags(q.Get("offers")),
		Wants:         util.ToSearchTags(q.Get("wants")),
		TaggedSince:   util.ParseTime(q.Get("tagged_since")),
		Category:      q.Get("category"),
		AccountNumber: q.Get("account_number"),
		City:          q.Get("city"),
		Region:        q.Get("region"),
		Country:       q.Get("country"),
		Balance:       balance,
		MaxPosBal:     maxPosBal,
		MaxNegBal:     maxNegBal,
	}

	return req, req.validate()
}

type AdminSearchEntityReq struct {
	Page          int
	PageSize      int
	Name          string
	Email         string
	Statuses      []string
	Offers        []string
	Wants         []string
	TaggedSince   time.Time
	Category      string
	City          string
	Region        string
	Country       string
	AccountNumber string
	Balance       *float64
	MaxPosBal     *float64
	MaxNegBal     *float64
}

func (req *AdminSearchEntityReq) validate() []error {
	errs := []error{}

	if !req.TaggedSince.IsZero() && len(req.Wants) == 0 && len(req.Offers) == 0 {
		errs = append(errs, errors.New("Please specify an offer or want tag."))
	}

	return errs
}

// GET /admin/entities/{entityID}

type AdminGetEntity struct {
	EntityID string
}

func NewAdminGetEntityReq(r *http.Request) (*AdminGetEntity, []error) {
	return &AdminGetEntity{
		EntityID: mux.Vars(r)["entityID"],
	}, nil
}

// PATCH /admin/entities/{entityID}

func NewAdminUpdateEntityReq(j AdminUpdateEntityJSON, originEntity *Entity, originBalanceLimit *BalanceLimit) (*AdminUpdateEntityReq, []error) {
	errs := j.validate()
	if len(errs) != 0 {
		return nil, errs
	}

	addedUsers := []string{}
	removedUsers := []string{}
	if j.Users != nil {
		addedUsers, removedUsers = util.StringDiff(*j.Users, util.ToIDStrings(originEntity.Users))
	}
	addedOffers := []string{}
	removedOffers := []string{}
	if j.Offers != nil {
		addedOffers, removedOffers = util.StringDiff(util.FormatTags(*j.Offers), TagFieldToNames(originEntity.Offers))
	}
	addedWants := []string{}
	removedWants := []string{}
	if j.Wants != nil {
		addedWants, removedWants = util.StringDiff(util.FormatTags(*j.Wants), TagFieldToNames(originEntity.Wants))
	}

	req := AdminUpdateEntityReq{
		OriginEntity:                       originEntity,
		OriginBalanceLimit:                 originBalanceLimit,
		Name:                               j.Name,
		Telephone:                          j.Telephone,
		Email:                              j.Email,
		IncType:                            j.IncType,
		CompanyNumber:                      j.CompanyNumber,
		Website:                            j.Website,
		DeclaredTurnover:                   j.DeclaredTurnover,
		Description:                        j.Description,
		ReceiveDailyMatchNotificationEmail: j.ReceiveDailyMatchNotificationEmail,
		ShowTagsMatchedSinceLastLogin:      j.ShowTagsMatchedSinceLastLogin,
		// Users
		Users:        j.Users,
		AddedUsers:   util.ToObjectIDs(addedUsers),
		RemovedUsers: util.ToObjectIDs(removedUsers),
		// Tags
		Offers:        j.Offers,
		AddedOffers:   addedOffers,
		RemovedOffers: removedOffers,
		Wants:         j.Wants,
		AddedWants:    addedWants,
		RemovedWants:  removedWants,
		Categories:    j.Categories,
		// Address
		Address:    j.Address,
		City:       j.City,
		Region:     j.Region,
		PostalCode: j.PostalCode,
		Country:    j.Country,
		// Account
		MaxPosBal: j.MaxPosBal,
		MaxNegBal: j.MaxNegBal,
		Status:    j.Status,
	}

	return &req, nil
}

type AdminUpdateEntityReq struct {
	OriginEntity                       *Entity
	OriginBalanceLimit                 *BalanceLimit
	Status                             string
	Name                               string
	Email                              string
	Telephone                          string
	IncType                            string
	CompanyNumber                      string
	Website                            string
	DeclaredTurnover                   *int
	Description                        string
	ReceiveDailyMatchNotificationEmail *bool
	ShowTagsMatchedSinceLastLogin      *bool
	// Users
	Users        *[]string
	AddedUsers   []primitive.ObjectID
	RemovedUsers []primitive.ObjectID
	// Tags
	Offers        *[]string
	AddedOffers   []string
	RemovedOffers []string
	Wants         *[]string
	AddedWants    []string
	RemovedWants  []string
	Categories    *[]string
	// Address
	Address    string
	City       string
	Region     string
	PostalCode string
	Country    string
	// Account
	MaxPosBal *float64
	MaxNegBal *float64
}

type AdminUpdateEntityJSON struct {
	Status           string    `json:"status"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Telephone        string    `json:"telephone"`
	IncType          string    `json:"incType"`
	CompanyNumber    string    `json:"companyNumber"`
	Website          string    `json:"website"`
	DeclaredTurnover *int      `json:"declaredTurnover"`
	Description      string    `json:"description"`
	Users            *[]string `json:"users"`
	// Tags
	Offers     *[]string `json:"offers"`
	Wants      *[]string `json:"wants"`
	Categories *[]string `json:"categories"`
	// Address
	Address    string `json:"address"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
	// flags
	ReceiveDailyMatchNotificationEmail *bool `json:"receiveDailyMatchNotificationEmail"`
	ShowTagsMatchedSinceLastLogin      *bool `json:"showTagsMatchedSinceLastLogin"`
	// Account
	MaxPosBal *float64 `json:"maxPositiveBalance"`
	MaxNegBal *float64 `json:"maxNegativeBalance"`
	// Useless (Do not use it)
	ID            string `json:"id"`
	AccountNumber string `json:"accountNumber"`
}

func (req *AdminUpdateEntityJSON) validate() []error {
	errs := []error{}

	if req.ID != "" {
		errs = append(errs, errors.New("The entity ID cannot be changed."))
	}
	if req.AccountNumber != "" {
		errs = append(errs, errors.New("The account number cannot be changed."))
	}
	if req.MaxPosBal != nil && *req.MaxPosBal < 0 {
		errs = append(errs, errors.New("The max positive balance should be positive."))
	}
	if req.MaxNegBal != nil && *req.MaxNegBal < 0 {
		errs = append(errs, errors.New("The max negative balance should be positive."))
	}

	categories := []string{}
	if req.Categories != nil {
		categories = *req.Categories
	}

	entity := Entity{
		Email:            req.Email,
		Name:             req.Name,
		Telephone:        req.Telephone,
		IncType:          req.IncType,
		CompanyNumber:    req.CompanyNumber,
		Website:          req.Website,
		DeclaredTurnover: req.DeclaredTurnover,
		Description:      req.Description,
		City:             req.City,
		Country:          req.Country,
		Address:          req.Address,
		Region:           req.Region,
		PostalCode:       req.PostalCode,
		Categories:       categories,
		Status:           req.Status,
	}
	errs = append(errs, entity.Validate()...)
	if req.Offers != nil {
		errs = append(errs, validateTags(*req.Offers)...)
	}
	if req.Wants != nil {
		errs = append(errs, validateTags(*req.Wants)...)
	}

	return errs
}

// DELETE /admin/entities/{entityID}

type AdminDeleteEntity struct {
	EntityID primitive.ObjectID
}

func NewAdminDeleteEntity(r *http.Request) (*AdminDeleteEntity, []error) {
	entityID := mux.Vars(r)["entityID"]
	if entityID == "" {
		return nil, []error{errors.New("Please enter entity id.")}
	}
	objectID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, []error{errors.New("Please enter valid entity id.")}
	}

	return &AdminDeleteEntity{
		EntityID: objectID,
	}, nil
}

// POST /admin/transfers

func NewAdminTransferReq(userReq *AdminTransferUserReq, payerEntity *Entity, payeeEntity *Entity) (*AdminTransferReq, []error) {
	req := &AdminTransferReq{
		PayerEntity:  payerEntity,
		PayeeEntity:  payeeEntity,
		TransferType: constant.TransferType.AdminTransfer,
		Amount:       userReq.Amount,
		Description:  userReq.Description,
	}
	return req, req.Validate()
}

type AdminTransferUserReq struct {
	Payer       string  `json:"payer"`
	Payee       string  `json:"payee"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type AdminTransferReq struct {
	PayerEntity  *Entity
	PayeeEntity  *Entity
	TransferType string // "Transfer" / "AdminTranser"
	Amount       float64
	Description  string
}

func (req *AdminTransferReq) Validate() []error {
	errs := []error{}

	// Amount should be positive value and with up to two decimal places.
	if req.Amount <= 0 || !util.IsDecimalValid(req.Amount) {
		errs = append(errs, errors.New("Please enter a valid numeric amount to send with up to two decimal places."))
	}

	// Only allow transfers with accounts that also have "trading-accepted" status
	if req.PayerEntity.Status != constant.Trading.Accepted {
		errs = append(errs, errors.New("Sender is not a trading member. Transfers can only be made when both entities have trading member status."))
	} else if req.PayeeEntity.Status != constant.Trading.Accepted {
		errs = append(errs, errors.New("Recipient is not a trading member. Transfers can only be made when both entities have trading member status."))
	}

	// Check if the user is doing the transaction to himself.
	if req.PayerEntity.AccountNumber == req.PayeeEntity.AccountNumber {
		errs = append(errs, errors.New("You cannot create a transaction with yourself."))
	}

	return errs
}

// GET /admin/transfers/{transferID}

func NewAdminGetTransfer(r *http.Request) (*AdminGetTransfer, []error) {
	req := &AdminGetTransfer{
		TransferID: mux.Vars(r)["transferID"],
	}
	return req, req.validate()
}

type AdminGetTransfer struct {
	TransferID string
}

func (req *AdminGetTransfer) validate() []error {
	errs := []error{}
	return errs
}

// GET /admin/transfers

func NewAdminSearchTransferQuery(r *http.Request) (*AdminSearchTransferReq, []error) {
	q := r.URL.Query()
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}
	dateFrom := util.ParseTime(q.Get("date_from"))
	dateTo := util.ParseTime(q.Get("date_to"))

	query := &AdminSearchTransferReq{
		Page:          page,
		PageSize:      pageSize,
		Offset:        (page - 1) * pageSize,
		Status:        getStatus(q.Get("status")),
		AccountNumber: q.Get("account_number"),
		DateFrom:      dateFrom,
		DateTo:        dateTo,
	}

	return query, query.validate()
}

type AdminSearchTransferReq struct {
	Page          int
	PageSize      int
	Offset        int
	Status        []string
	AccountNumber string
	DateFrom      time.Time
	DateTo        time.Time
}

func (req *AdminSearchTransferReq) validate() []error {
	errs := []error{}
	for _, s := range req.Status {
		if s != "initiated" && s != "completed" && s != "cancelled" {
			errs = append(errs, errors.New("Please specify valid status."))
		}
	}
	return errs
}

func getStatus(input string) []string {
	splitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}
	return strings.FieldsFunc(strings.ToLower(input), splitFn)
}

// GET /admin/logs

func NewAdminSearchLog(r *http.Request) (*AdminSearchLogReq, []error) {
	q := r.URL.Query()
	page, err := util.ToInt(q.Get("page"), 1)
	if err != nil {
		return nil, []error{err}
	}
	pageSize, err := util.ToInt(q.Get("page_size"), viper.GetInt("page_size"))
	if err != nil {
		return nil, []error{err}
	}
	dateFrom := util.ParseTime(q.Get("date_from"))
	dateTo := util.ParseTime(q.Get("date_to"))

	query := &AdminSearchLogReq{
		Page:       page,
		PageSize:   pageSize,
		Offset:     (page - 1) * pageSize,
		Email:      q.Get("email"),
		Categories: getCategories(q.Get("category")),
		Action:     q.Get("action"),
		Detail:     q.Get("detail"),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}

	return query, query.validate()
}

type AdminSearchLogReq struct {
	Page       int
	PageSize   int
	Offset     int
	Email      string
	Categories []string
	Action     string
	Detail     string
	DateFrom   time.Time
	DateTo     time.Time
}

func (req *AdminSearchLogReq) validate() []error {
	errs := []error{}
	for _, c := range req.Categories {
		if c != "user" && c != "admin" {
			errs = append(errs, errors.New("Please specify valid status."))
		}
	}
	return errs
}

func getCategories(input string) []string {
	splitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}
	return strings.FieldsFunc(strings.ToLower(input), splitFn)
}
