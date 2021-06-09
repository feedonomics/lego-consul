package cmd

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/hashicorp/consul/api"
	"github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/feedonomics/lego-consul/config"
	"github.com/feedonomics/lego-consul/paths"
	"github.com/feedonomics/lego-consul/solvers"
	"github.com/feedonomics/lego-consul/types"
	"github.com/feedonomics/lego-consul/utility"
)

var (
	forceRenewal bool
	postHookCmd  string
)

var renewCmd = &cobra.Command{
	Use:   `renew`,
	Short: `Renew all previously obtained certificates that are near expiry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		RenewalCfgs, err := filepath.Glob(config.Path + `/renewal/*.json`)
		if err != nil {
			return err
		}

		if len(RenewalCfgs) == 0 {
			logrus.Infof(`no renewal configurations found`)
		}

		var RequirePostHook bool
		for _, Config := range RenewalCfgs {
			HaveRenewal, err := ProcessRenewal(Config, forceRenewal)
			if err != nil {
				return err
			}
			if HaveRenewal {
				RequirePostHook = true
			}
		}

		if postHookCmd != `` && RequirePostHook {
			return RunPostHook(postHookCmd)
		}
		return nil
	},
}

func GetSolver(Cfg types.Renewal) (challenge.Provider, error) {
	switch Cfg.Authenticator {
	case `consul`:
		ConsulAgent, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return nil, err
		}

		return &solvers.ConsulSolver{Agent: ConsulAgent}, nil
	default:
		return nil, fmt.Errorf(`solvers: unsupported renewal authenticator '%s'`, Cfg.Authenticator)
	}
}

func ProcessRenewal(RenewalCfgPath string, Force bool) (bool, error) {
	var RenewalCfg types.Renewal
	Contents, err := ioutil.ReadFile(RenewalCfgPath)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(Contents, &RenewalCfg); err != nil {
		return false, err
	}

	Challenge, err := GetSolver(RenewalCfg)
	if err != nil {
		return false, err
	}

	Certificates, err := readCertificate(RenewalCfg.Paths.Certificate)
	if err != nil {
		return false, err
	}

	if len(Certificates) == 0 {
		return false, fmt.Errorf(`[%s] no certificates loaded from Certificate file path`, RenewalCfg.Domain)
	}

	if !Force && !needRenewal(Certificates[0], RenewalCfg.Domain, RenewalCfg.SANs, 30) {
		logrus.Infof(`[%s] certificate renewal is not required`, RenewalCfg.Domain)
		return false, nil
	}

	logrus.Infof(`[%s] starting renewal process`, RenewalCfg.Domain)

	User, err := types.LoadAccount(RenewalCfg.Server, RenewalCfg.Account)
	if err != nil {
		return false, err
	}

	LegoConfig := lego.NewConfig(User)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	LegoConfig.CADirURL = RenewalCfg.Server
	LegoConfig.Certificate.KeyType = certcrypto.RSA2048
	if config.TLSInsecure {
		LegoConfig.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
	}

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(LegoConfig)
	if err != nil {
		return false, err
	}

	_ = client.Challenge.SetHTTP01Provider(Challenge)

	Domains := []string{RenewalCfg.Domain}
	Domains = append(Domains, RenewalCfg.SANs...)

	request := certificate.ObtainRequest{
		Domains: Domains,
		Bundle:  false,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return false, err
	}

	archiveFilePaths, err := paths.GetArchiveFileSet(certificates.Domain)
	if err != nil {
		return false, err
	}

	if err := archiveFilePaths.WriteFiles(certificates); err != nil {
		return false, err
	}

	liveSet := paths.GetLiveFileSet(certificates.Domain)
	if err := archiveFilePaths.Activate(liveSet); err != nil {
		return false, err
	}

	return true, nil
}

func RunPostHook(v string) error {
	logrus.Infof(`[Post-Hook] [Cmd]: %s`, v)
	split, err := shellquote.Split(v)
	if err != nil {
		return fmt.Errorf(`post-hook: %s`, err)
	}
	var args []string
	if len(split) > 1 {
		args = split[1:]
	}

	StdOut := bytes.Buffer{}
	StdErr := bytes.Buffer{}

	Cmd := exec.Command(split[0], args...)
	Cmd.Stdout = &StdOut
	Cmd.Stderr = &StdErr
	err = Cmd.Run()

	StdOutB := StdOut.Bytes()
	StdErrB := StdErr.Bytes()

	if len(StdOutB) > 0 {
		logrus.Infof(`[Post-Hook] [stdout] %s`, StdOutB)
	}

	if len(StdErrB) > 0 {
		logrus.Errorf(`[Post-Hook] [stderr] %s`, StdErrB)
	}

	if err != nil {
		return fmt.Errorf(`[Post-Hook] Command Error: %s`, err)
	}

	return nil
}

func readCertificate(path string) ([]*x509.Certificate, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// The input may be a bundle or a single certificate.
	return certcrypto.ParsePEMBundle(content)
}

func needRenewal(x509Cert *x509.Certificate, domain string, AltNames []string, days int) bool {
	if x509Cert.IsCA {
		logrus.Fatalf("[%s] Certificate bundle starts with a CA certificate", domain)
	}

	for _, desiredAltName := range AltNames {
		if desiredAltName != domain &&
			!utility.ExistsStrings(x509Cert.DNSNames, desiredAltName) {
			logrus.Infof(`[%s] The certificate does not include the requested subject alternate name %s: yes renewal`,
				domain, desiredAltName)
			return true
		}
	}

	if days >= 0 {
		notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
		if notAfter > days {
			logrus.Printf("[%s] The certificate expires in %d days, the number of days defined to perform the renewal is %d: no renewal.",
				domain, notAfter, days)
			return false
		}
	}

	return true
}
