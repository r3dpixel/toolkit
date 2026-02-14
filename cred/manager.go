package cred

const (
	userKey   = "username"
	secretKey = "password"
)

// Mode the mode in which a IdentityManager can function
type Mode byte

const (
	KeyRing Mode = iota // IdentityManager will use the OS keyring
	Env                 // IdentityManager will use the environment through environment variables
)

// manager internally uses IdentityProvider to read/write credentials
type manager struct {
	provider IdentityProvider
}

// NewManager creates a new identity manager with the specified label and Mode
func NewManager(credLabel string, mode Mode) IdentityManager {
	return &manager{
		provider: getProvider(credLabel, mode),
	}
}

// SetAll sets both user and secret credentials from the provided identity
func (m *manager) SetAll(identity Identity) error {
	if err := m.provider.Set(userKey, identity.User); err != nil {
		return err
	}
	return m.provider.Set(secretKey, identity.Secret)
}

// Set updates credentials based on the provided payload, setting only non-nil values
func (m *manager) Set(payload IdentityPayload) error {
	var err error

	if payload.User != nil {
		err = m.provider.Set(userKey, *payload.User)
	}

	if err == nil && payload.Secret != nil {
		err = m.provider.Set(secretKey, *payload.Secret)
	}

	return err
}

// SetUser sets the username credential
func (m *manager) SetUser(user string) error {
	return m.provider.Set(userKey, user)
}

// SetSecret sets the secret credential
func (m *manager) SetSecret(secret string) error {
	return m.provider.Set(secretKey, secret)
}

// Get retrieves both user and secret credentials and returns them as an Identity
func (m *manager) Get() (Identity, error) {
	user, err := m.provider.Get(userKey)
	if err != nil {
		return Identity{}, err
	}
	secret, err := m.provider.Get(secretKey)
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		User:   user,
		Secret: secret,
	}, nil
}

// GetUser retrieves the username credential
func (m *manager) GetUser() (string, error) {
	return m.provider.Get(userKey)
}

// GetSecret retrieves the password credential
func (m *manager) GetSecret() (string, error) {
	return m.provider.Get(secretKey)
}

// Delete removes both user and secret credentials from storage
func (m *manager) Delete() error {
	if err := m.provider.Delete(userKey); err != nil {
		return err
	}
	return m.provider.Delete(secretKey)
}

// CredLabel returns the label of the identity provider
func (m *manager) CredLabel() string {
	return m.provider.CredLabel()
}

// getProvider returns the appropriate identity provider based on the specified mode
func getProvider(label string, mode Mode) IdentityProvider {
	switch mode {
	case Env:
		return NewEnvProvider(label)
	case KeyRing:
		return NewKeyProvider(label)
	}
	return nil
}
