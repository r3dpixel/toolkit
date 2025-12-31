package cred

// envProvider implements IdentityProvider using environment variables
type envProvider struct {
	credLabel string
}

// NewEnvProvider creates a new environment variable based identity provider
func NewEnvProvider(credLabel string) IdentityProvider {
	return &envProvider{
		credLabel: credLabel,
	}
}

// Set stores a key-value pair in the environment variables
func (p *envProvider) Set(key, value string) error {
	return ToEnv(p.credLabel, key, value)
}

// Get retrieves a value for the given key from environment variables
func (p *envProvider) Get(key string) (string, error) {
	return FromEnv(p.credLabel, key)
}

// Delete removes a key-value pair from the environment variables
func (p *envProvider) Delete(key string) error {
	return DeleteEnv(p.credLabel, key)
}

// CredLabel returns the label for the provider
func (p *envProvider) CredLabel() string {
	return p.credLabel
}
