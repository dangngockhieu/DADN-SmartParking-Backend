package token

type ClaimsPayload struct {
	UserID uint
	Email  string
	Role   string
	JTI    string
	Exp    int64
}
