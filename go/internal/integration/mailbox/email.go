package mailbox

type Email struct {
	ID         string `json:"id"`
	Account    string `json:"account"`
	From       string `json:"from"`
	FromName   string `json:"fromName"`
	Subject    string `json:"subject"`
	Preview    string `json:"preview"`
	ReceivedAt string `json:"receivedAt"`
}
