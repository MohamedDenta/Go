package service

//ResponseCreateVoucher create voucher response
type ResponseCreateVoucher struct {
	ID          string `json:"_id"`
	Create_time int64  `json:"create_time"`
	Used        int    `json:"used"`
	Status      string `json:"status"`
	Code        string `json:"code"`
}

//ServCredentials service credentials
type ServCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Voucher voucher structure
type Voucher struct {
	Cmd    string `json:"cmd"`
	N      string `json:"n"`
	Expire string `json:"expire"`
	Up     string `json:"up"`
	Down   string `json:"down"`
	Bytes  string `json:"bytes"`
	MBytes string `json:"MBytes"`
	Note   string `json:"note"`
	Quota  string `json:"quota"`
}

// UserKeys ..
type UserKeys struct {
	PublicKey string
	Password  string
}

//VoucherResponse voucher response
type VoucherResponse struct {
	TransactionId string
	Voucher       string
}

//VoucherRequest voucher request
type VoucherRequest struct {
	Cmd    string
	N      string
	expire int
	up     string
	down   string
	bytes  string
	MBytes bool
	note   string
	quota  string
}

//Vocherstatus ...
type Vocherstatus struct {
	PublicKey string
	Password  string
	VoucherID string
}

//VocherstatusResponse ...
type VocherstatusResponse struct {
	Id          string
	Site_id     string
	Create_time int64
}
