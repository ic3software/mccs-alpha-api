package constant

// Entity status
var Entity = struct {
	Pending  string
	Accepted string
	Rejected string
}{
	Pending:  "pending",
	Accepted: "accepted",
	Rejected: "rejected",
}

// Trading Status decides whether a entity can perform transactions (already in accepted status).
var Trading = struct {
	Pending  string
	Accepted string
	Rejected string
}{
	Pending:  "tradingPending",
	Accepted: "tradingAccepted",
	Rejected: "tradingRejected",
}
