package apierr

// Instantiate / HTTP error codes (see reference.md).

const (
	CodeMethodNotAllowed = 10001
	CodeUnauthorized     = 10002

	CodeInvalidJSON       = 10101
	CodeJSONUnknownField  = 10102

	CodeMissingEmail         = 10201
	CodeMissingContractUID   = 10202
	CodeMissingCompany       = 10203
	CodeMissingProject       = 10204
	CodeMissingWebhookURL    = 10205
	CodeMissingAuthorization = 10206

	CodePersistFailed = 10901
)
