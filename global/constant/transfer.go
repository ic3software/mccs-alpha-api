package constant

var Transfer = struct {
	Initiated string
	Completed string
	Cancelled string
}{
	Initiated: "transferInitiated",
	Completed: "transferCompleted",
	Cancelled: "transferCancelled",
}

var TransferType = struct {
	In  string
	Out string
}{
	In:  "in",
	Out: "out",
}
