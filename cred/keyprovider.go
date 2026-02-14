package cred

import "sync"

// keyProvider implements IdentityProvider using the OS keyring
type keyProvider struct {
	mu        sync.Mutex
	credLabel string
}

// NewKeyProvider creates a new key provider with the specified label.
func NewKeyProvider(credLabel string) IdentityProvider {
	return &keyProvider{
		credLabel: credLabel,
	}
}

// Set stores a key-value pair in the keyring with thread-safe access.
func (p *keyProvider) Set(key, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return ToKeyRing(p.credLabel, key, value)
}

// Get retrieves a value by key from the keyring with thread-safe access.
func (p *keyProvider) Get(key string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return FromKeyRing(p.credLabel, key)
}

// Delete removes a key from the keyring with thread-safe access.
func (p *keyProvider) Delete(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return DeleteKeyRing(p.credLabel, key)
}

// CredLabel returns the label for the provider
func (p *keyProvider) CredLabel() string {
	return p.credLabel
}
