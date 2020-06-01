package serviceModule

import "../transactionModule"

// InquiryResponse ..
type InquiryResponse struct {
	ID     string
	Amount string
	Msg    string
}

// PurchaseResponse ..
type PurchaseResponse struct {
	ID  string
	Msg string
}

// PurchaseServiceStruct ..
type PurchaseServiceStruct struct {
	ID             string
	Password       string
	Transactionobj transactionModule.DigitalwalletTransaction
}
