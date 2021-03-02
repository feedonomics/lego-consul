package solvers

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

type ConsulSolver struct {
	Agent *api.Client
}

func (c ConsulSolver) Present(domain, token, keyAuth string) error {
	logrus.Infof(`[INFO] [%s] consul: storing challenge token %s`, domain, token)
	KeyPath := fmt.Sprintf(`acme-challenges/%s/%s`, domain, token)
	_, err := c.Agent.KV().Put(&api.KVPair{
		Key:   KeyPath,
		Value: []byte(keyAuth),
	}, nil)
	return err
}

func (c ConsulSolver) CleanUp(domain, token, _ string) error {
	logrus.Infof(`[INFO] [%s] consul: removing challenge token %s`, domain, token)
	KeyPath := fmt.Sprintf(`acme-challenges/%s/%s`, domain, token)
	_, err := c.Agent.KV().Delete(KeyPath, nil)
	return err
}
