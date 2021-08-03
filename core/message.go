package core

type Message struct {
	Source      RSymbol
	Destination RSymbol
	Reason      Reason
	Content     interface{}
}

type Reason string

const (
	NewTransInfos              = Reason("NewTransInfo")
	ReportTransResultWithBlock = Reason("ReportTransResultWithBlock")
	ReportTransResultWithIndex = Reason("ReportTransResultWithIndex")
	SubmitSignature            = Reason("SubmitSignature")
	SignatureEnough            = Reason("SignatureEnough")
	GetLatestDealBLock         = Reason("GetLatestDealBLock")
	GetSignatures              = Reason("GetSignatures")
)
