package cmd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"

	// "github.com/julienschmidt/httprouter"
	"net/http"

	"github.com/spf13/cobra"
)

var httpConsulBind string

var httpConsulCmd = &cobra.Command{
	Use:   `http-consul`,
	Short: `HTTP Challenge Responder via Consul.io KeyValue store.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		Agent, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return err
		}

		Mux := httprouter.New()
		Mux.GET(`/.well-known/acme-challenge/:token`, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			Token := p.ByName(`token`)

			// TODO: Handle X-ForwardedFor is Reverse Proxy.
			HostParts := strings.SplitN(r.Host, `:`, 2)
			HostOnly := HostParts[0]

			logrus.Infof(`[%s] received request for challenge token: %s`, HostOnly, Token)

			KeyPath := fmt.Sprintf(`/acme-challenges/%s/%s`, HostOnly, Token)
			logrus.Debugf(`[%s] Key Path: %s`, HostOnly, KeyPath)

			Value, _, err := Agent.KV().Get(KeyPath, nil)
			if err != nil {
				logrus.Warnf(`[%s] error querying for key auth: %s`, HostOnly, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if Value == nil {
				logrus.Warnf(`[%s] error querying for key auth: Key Not Found`, HostOnly)
				http.Error(w, `404 page not found`, http.StatusNotFound)
				return
			}

			// successful
			w.Header().Set("Content-Type", "text/plain")
			_, err = w.Write(Value.Value)
			if err != nil {
				logrus.Warnf(`[%s] error writing key auth response: %s`, HostOnly, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logrus.Infof("[%s] Served key authentication", HostOnly)
		})

		logrus.Infof(`[HTTP] Start listening on %s`, httpConsulBind)
		return http.ListenAndServe(httpConsulBind, Mux)
	},
}
