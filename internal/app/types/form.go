package types

type RegisterData struct {
	User             *User
	Entity           *EntityData
	ConfirmPassword  string
	ConfirmEmail     string
	Terms            string
	RecaptchaSitekey string
}

type UpdateAccountData struct {
	User            *User
	Entity          *EntityData
	Balance         *BalanceLimit
	CurrentPassword string
	ConfirmPassword string
}
