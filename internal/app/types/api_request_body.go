package types

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

type UpdateUserReqBody struct {
	ID                            string `json:"id"`
	Email                         string `json:"email"`
	FirstName                     string `json:"firstName"`
	LastName                      string `json:"lastName"`
	UserPhone                     string `json:"userPhone"`
	DailyEmailMatchNotification   *bool  `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin *bool  `json:"showTagsMatchedSinceLastLogin"`
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
