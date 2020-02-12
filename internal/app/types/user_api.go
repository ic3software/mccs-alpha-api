package types

type SignupRequest struct {
	Email                 string `json:"email"`
	Password              string `json:"password"`
	FirstName             string `json:"firstName"`
	LastName              string `json:"lastName"`
	UserPhone             string `json:"userPhone"`
	ShowRecentMatchedTags *bool  `json:"showRecentMatchedTags"`
	DailyNotification     *bool  `json:"DailyNotification"`
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

type UpdateUser struct {
	FirstName                     string `json:"firstName"`
	LastName                      string `json:"lastName"`
	UserPhone                     string `json:"userPhone"`
	DailyEmailMatchNotification   *bool  `json:"dailyEmailMatchNotification"`
	ShowTagsMatchedSinceLastLogin *bool  `json:"showTagsMatchedSinceLastLogin"`
}

type UpdateUserEntity struct {
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
