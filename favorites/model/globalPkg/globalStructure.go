package globalPkg

//JSONString take string
type JSONString struct {
	Name string
}

/*----------------global variable structure----------------- */
type GlobalVariables struct {
	ValidatorFees                  float64
	ElectionTimeInSeconds          int32
	NoOfTransactionsPerBlock       int
	TransactionStakeCoins          float64
	StringFixedLength              int
	DeleteAccountLoopTimeInseacond int32
	DeleteAccountTimeInseacond     float64
	MaxConfirmcode                 int
	TxValidationTimeInSeconds      int
	TokenIDStringFixedLength       int
	TransactionFee                 float64
	TransactionRefundFee           float64
	InoCoinToDollarRatio           float64
	ServiceLogin                   string
	ServiceCreateVoutcher          string
	ServiceStatus                  string
	RSAKeyBitSize                  int
	Downloadfileip                 string
}

type GlobalServerIp struct {
	Ip   string
	Port string
}
type SendMessage struct {
	MessageAPI string
}

var IsDown = false

//var AdminObj AdminStruct
var GlobalObj GlobalVariables
var GlobalServerObj GlobalServerIp

/*----------------validation-----*/
func Validation(globalObj GlobalVariables) bool {
	return true
}

type ResponseCreateVoucher struct {
	ID          string `json:"_id"`
	Create_time int64  `json:"create_time"`
	Used        int    `json:"used"`
	Status      string `json:"status"`
	Code        string `json:"code"`
}
type ServCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
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

var CookieObject2 []string

var UserSigningKey = []byte("ZGVudGEuY29t")
var AdminSigningKey = []byte("endlcnQuY29t")

//AESEncryptionKey is the key used to encrypt ECC keys with befor saving on DB
var AESEncryptionKey = "This is phase to encrypt Public ckey with..."

//RSAPublic rsa public key
var RSAPublic string

//RSAPrivate rsa private key
var RSAPrivate string
