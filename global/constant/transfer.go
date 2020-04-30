package constant

func MapTransferType(name string) string {
	if name == "initiated" {
		return Transfer.Initiated
	} else if name == "completed" {
		return Transfer.Completed
	} else if name == "cancelled" {
		return Transfer.Cancelled
	}
	return "unknown"
}

var Transfer = struct {
	Initiated string
	Completed string
	Cancelled string
}{
	Initiated: "transferInitiated",
	Completed: "transferCompleted",
	Cancelled: "transferCancelled",
}

var TransferDirection = struct {
	In  string
	Out string
}{
	In:  "in",
	Out: "out",
}

var TransferType = struct {
	Transfer      string
	AdminTransfer string
}{
	Transfer:      "transfer",
	AdminTransfer: "adminTransfer",
}
