package config

import "github.com/go-acme/lego/v4/lego"

var (
	DirectoryURL = lego.LEDirectoryProduction
	Path         string
	TLSInsecure  bool
	Quiet        bool
)
