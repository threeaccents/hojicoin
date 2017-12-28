package hoji

// General errors.
const (
	ErrUnauthorized     = Error("unauthorized")
	ErrInternal         = Error("internal error")
	ErrNotFound         = Error("resource not found")
	ErrBadRequest       = Error("bad request")
	ErrBucketNotExist   = Error("bucket does not exist")
	ErrInsuficientFunds = Error("insufficient funds")
)

// Error represents a Vano error.
type Error string

// Error returns the error message.
func (e Error) Error() string {
	return string(e)
}
