package cred

// Identity stores a pair of user - secret
type Identity struct {
	User   string
	Secret string
}

// IdentityPayload stores a pair of user - secret where any could be missing
type IdentityPayload struct {
	User   *string
	Secret *string
}

// IdentityProvider generic API for identity storage
type IdentityProvider interface {
	CredLabel() string
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// IdentityReader generic API for reading credentials
type IdentityReader interface {
	CredLabel() string
	Get() (Identity, error)
	GetUser() (string, error)
	GetSecret() (string, error)
}

// IdentityWriter generic API for writing credentials
type IdentityWriter interface {
	CredLabel() string
	SetAll(identity Identity) error
	Set(payload IdentityPayload) error
	SetUser(user string) error
	SetSecret(secret string) error
	Delete() error
}

// IdentityManager generic API for reading/writing credentials
type IdentityManager interface {
	IdentityReader
	IdentityWriter
}
