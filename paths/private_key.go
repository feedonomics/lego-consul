package paths

import (
	"crypto"
	"fmt"
	"io/ioutil"

	"gopkg.in/square/go-jose.v2"
)

func LoadPrivateKey(Path string) (crypto.PrivateKey, error) {
	Content, err := ioutil.ReadFile(Path)
	if err != nil {
		return nil, err
	}

	Key := jose.JSONWebKey{}
	if err := Key.UnmarshalJSON(Content); err != nil {
		return nil, err
	}

	if PrivateKey, ok := Key.Key.(crypto.PrivateKey); ok {
		return PrivateKey, nil
	}

	return nil, fmt.Errorf(`accounts: no private key found in %s`, Path)
}
