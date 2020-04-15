package cnb

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	ggcr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Registry struct {
	Keychain authn.Keychain
}

func (r *Registry) GetImage(ref string) (ggcr.Image, error) {
	parsed, err := name.ParseReference(ref, name.WeakValidation)
	if err != nil {
		return nil, err
	}
	// TODO add an LRU cache to avoid slamming the remote registry with requests
	return remote.Image(parsed, remote.WithAuthFromKeychain(r.Keychain))
}
