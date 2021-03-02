package cmd

import (
	"fmt"
	"net/http"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ottodashadow/lego-consul/config"
	"github.com/ottodashadow/lego-consul/paths"
	"github.com/ottodashadow/lego-consul/solvers"
	"github.com/ottodashadow/lego-consul/types"
	"github.com/ottodashadow/lego-consul/utility"
)

var generateCmd = &cobra.Command{
	Use:   `generate`,
	Short: `Generate a new certificate.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ConsulAgent, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return err
		}

		User := types.User{
			Email: Email,
		}

		//if err := User.GenerateEllipticCurveKey(elliptic.P256()); err != nil {
		if err := User.GenerateRSAKey(2048); err != nil {
			return err
		}

		AccountID, err := User.AccountID()
		if err != nil {
			return err
		}

		LegoConfig := lego.NewConfig(&User)

		// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
		LegoConfig.CADirURL = config.DirectoryURL
		LegoConfig.Certificate.KeyType = certcrypto.RSA2048
		if config.TLSInsecure {
			LegoConfig.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
		}

		// A client facilitates communication with the CA server.
		client, err := lego.NewClient(LegoConfig)
		if err != nil {
			return err
		}

		// New users will need to register
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		User.Registration = reg

		_ = client.Challenge.SetHTTP01Provider(&solvers.ConsulSolver{Agent: ConsulAgent})

		SANs := utility.ParseSANs(SANs)
		Domains := []string{Domain}
		Domains = append(Domains, SANs...)

		request := certificate.ObtainRequest{
			Domains: Domains,
			Bundle:  false,
		}
		certificates, err := client.Certificate.Obtain(request)
		if err != nil {
			return err
		}

		if err := User.WriteFiles(config.DirectoryURL); err != nil {
			return err
		}

		archiveFilePaths, err := paths.GetArchiveFileSet(certificates.Domain)
		if err != nil {
			return err
		}

		if err := archiveFilePaths.WriteFiles(certificates); err != nil {
			return err
		}

		liveSet := paths.GetLiveFileSet(certificates.Domain)
		if err := archiveFilePaths.Activate(liveSet); err != nil {
			return err
		}

		renewal := types.Renewal{
			Account:       AccountID,
			Authenticator: `consul`,
			Server:        config.DirectoryURL,
			Domain:        Domain,
			SANs:          SANs,
			Paths: types.RenewalPaths{
				Certificate: liveSet.Certificate,
				Chain:       liveSet.Chain,
				FullChain:   liveSet.FullChain,
				PrivateKey:  liveSet.PrivateKey,
			},
		}

		// write account's regr.json file.
		logrus.Infof(`[INFO] [%s] renewal: writing renewal configuration`, certificates.Domain)
		renewalFile := fmt.Sprintf(`%s/renewal/%s.json`, config.Path, certificates.Domain)
		if err := paths.WriteFileJSON(renewalFile, renewal, 0644); err != nil {
			return err
		}

		return nil
	},
}
