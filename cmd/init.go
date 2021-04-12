package cmd

import (
	"os"

	"github.com/go-acme/lego/v4/lego"

	"github.com/ottodashadow/lego-consul/config"
	"github.com/ottodashadow/lego-consul/version"
)

var (
	Domain string
	SANs   []string
	Email  string
)

func init() {
	defaultPath := os.Getenv(`LEGO_PATH`)
	if defaultPath == `` {
		defaultPath = `/etc/lego-acme`
	}

	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true

	RootCmd.AddCommand(httpConsulCmd)
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(renewCmd)
	RootCmd.AddCommand(manageSANsCmd)

	manageSANsCmd.AddCommand(addSANsCmd)

	RootCmd.PersistentFlags().BoolVar(&config.Quiet, `quiet`, false, `Silence all output except errors. Useful for automation via cron.`)
	RootCmd.PersistentFlags().BoolVar(&config.Syslog, `syslog`, false, `Write log events to syslog instead of stderr.`)

	generateCmd.Flags().StringVar(&config.DirectoryURL, `server`, lego.LEDirectoryProduction, `CA Server Directory URL`)
	generateCmd.Flags().BoolVar(&config.TLSInsecure, `insecure`, false, `Skip TLS Verification of ACME Server (use for local dev)`)
	generateCmd.Flags().StringVar(&Domain, `domain`, ``, `Main domain name to generate a certificate for. Should be servers official hostname.`)
	generateCmd.Flags().StringArrayVar(&SANs, `sans`, []string{}, `Subject Alt-Names to include when generating the certificate.`)
	generateCmd.Flags().StringVar(&Email, `email`, ``, `Email used for registration and recovery contact.`)
	generateCmd.Flags().StringVar(&config.Path, `path`, defaultPath, `Directory to use for storing the data.`)

	renewCmd.Flags().StringVar(&config.DirectoryURL, `server`, lego.LEDirectoryProduction, `CA Server Directory URL`)
	renewCmd.Flags().BoolVar(&config.TLSInsecure, `insecure`, false, `Skip TLS Verification of ACME Server (use for local dev)`)
	renewCmd.Flags().BoolVar(&forceRenewal, `force`, false, `Ignore certificate expiration and force generation of new certificates`)
	renewCmd.Flags().StringVar(&postHookCmd, `post-hook`, ``, `Command to run after certificate renewal process.`)

	addSANsCmd.Flags().StringVar(&Domain, `domain`, ``, `Main domain name to generate a certificate for. Should be servers official hostname.`)
	addSANsCmd.Flags().StringArrayVar(&SANs, `sans`, []string{}, `Subject Alt-Names to include when generating the certificate.`)

	httpConsulCmd.Flags().StringVar(&httpConsulBind, `bind`, `127.0.0.1:5002`, `Address and port to listen on.`)

	RootCmd.Flags().BoolP("version", "v", false, "")
	RootCmd.Version = version.String()
	RootCmd.SetVersionTemplate(`{{printf "LEGO Consul "}}{{printf "%s" .Version}}
`)
}
