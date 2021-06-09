package types

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"

	"github.com/go-acme/lego/v4/registration"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2"

	"github.com/feedonomics/lego-consul/paths"
)

var (
	ErrUserNoPublicKey = errors.New(`accounts: User key is not a crypto.PublicKey`)
)

// You'll need a user or account type that implements acme.User
type User struct {
	Email        string
	Registration *registration.Resource
	// private load via JSON/PEM files.
	key crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (u *User) GenerateEllipticCurveKey(Curve elliptic.Curve) error {
	Key, err := ecdsa.GenerateKey(Curve, rand.Reader)
	if err != nil {
		return err
	}
	u.key = Key
	return nil
}

func (u *User) GenerateRSAKey(bits int) error {
	Key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	u.key = Key
	return nil
}

func (u *User) GetJSONWebKey() jose.JSONWebKey {
	return jose.JSONWebKey{
		Key: u.key,
	}
}

type publicKeyProvider interface {
	Public() crypto.PublicKey
}

func (u *User) AccountID() (string, error) {
	if PublicProvider, ok := u.key.(publicKeyProvider); ok {
		publicBytes, err := x509.MarshalPKIXPublicKey(PublicProvider.Public())
		if err != nil {
			return "", err
		}
		publicPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicBytes,
		})

		Hash := md5.New()
		_, _ = Hash.Write(publicPEM)
		Sum := Hash.Sum(nil)

		return hex.EncodeToString(Sum), nil
	}
	return ``, ErrUserNoPublicKey
}

func (u *User) WriteFiles(DirectoryURL string) error {
	AccountID, err := u.AccountID()
	if err != nil {
		return err
	}

	AccountFolder, err := paths.GetAccountFolder(DirectoryURL, AccountID)
	if err != nil {
		return err
	}

	// write account's private_key.json file.
	logrus.Infof(`[INFO] account: writing account private_key.json`)
	if err := paths.WriteFileJSON(AccountFolder+`/private_key.json`, u.GetJSONWebKey(), 0400); err != nil {
		return err
	}

	// write account's regr.json file.
	logrus.Infof(`[INFO] account: writing account regr.json`)
	if err := paths.WriteFileJSON(AccountFolder+`/regr.json`, u.GetRegistration(), 0644); err != nil {
		return err
	}

	return nil
}

func LoadAccount(DirectoryURL string, AccountID string) (*User, error) {
	AccountFolder, err := paths.GetAccountFolder(DirectoryURL, AccountID)
	if err != nil {
		return nil, err
	}

	reg := registration.Resource{}
	logrus.Infof(`[Account] [%s] reading regr.json configuration`, AccountID)
	if err := paths.ReadFileJSON(AccountFolder+`/regr.json`, &reg); err != nil {
		return nil, err
	}

	logrus.Infof(`[Account] [%s] reading private_key.json crypto key`, AccountID)
	Key, err := paths.LoadPrivateKey(AccountFolder + `/private_key.json`)
	if err != nil {
		return nil, err
	}

	return &User{
		Registration: &reg,
		key:          Key,
	}, nil
}
