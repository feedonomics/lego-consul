package paths

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/sirupsen/logrus"

	"github.com/ottodashadow/lego-consul/config"
)

type FilePathSet struct {
	Domain      string
	Directory   string
	Certificate string
	PrivateKey  string
	Chain       string
	FullChain   string
}

func GetAccountFolder(CAURL string, AccountID string) (string, error) {
	parsed, err := url.Parse(CAURL)
	if err != nil {
		return ``, err
	}

	CleanHost := strings.Replace(parsed.Host, `:`, `-`, 1)
	return filepath.Clean(config.Path + `/accounts/` + CleanHost + `/` + parsed.Path + `/` + AccountID), nil
}

func GetArchiveFileSet(Domain string) (FilePathSet, error) {
	archiveFolder := path.Clean(config.Path + `/archive/` + Domain)
	index := 1
	for {
		Set := FilePathSet{
			Domain:      Domain,
			Directory:   path.Clean(archiveFolder),
			Certificate: path.Clean(fmt.Sprintf(`%s/cert%d.pem`, archiveFolder, index)),
			PrivateKey:  path.Clean(fmt.Sprintf(`%s/privkey%d.pem`, archiveFolder, index)),
			Chain:       path.Clean(fmt.Sprintf(`%s/chain%d.pem`, archiveFolder, index)),
			FullChain:   path.Clean(fmt.Sprintf(`%s/fullchain%d.pem`, archiveFolder, index)),
		}
		files := []string{Set.Certificate, Set.PrivateKey, Set.Chain, Set.FullChain}

		found := false
		for _, f := range files {
			_, err := os.Stat(f)
			if err == nil {
				// file was found continue
				found = true
			}

			// err returned but not a not found error
			if err != nil && !os.IsNotExist(err) {
				return FilePathSet{}, err
			}
		}

		if !found {
			return Set, nil
		}

		index++
	}
}

func GetLiveFileSet(Domain string) FilePathSet {
	Directory := path.Clean(config.Path + `/live/` + Domain)
	return FilePathSet{
		Domain:      Domain,
		Directory:   Directory,
		Certificate: Directory + `/cert.pem`,
		PrivateKey:  Directory + `/privkey.pem`,
		Chain:       Directory + `/chain.pem`,
		FullChain:   Directory + `/fullchain.pem`,
	}
}

func (Set *FilePathSet) WriteFiles(certificates *certificate.Resource) error {
	logrus.Infof(`[INFO] [%s] files: Writing archive certificate files.`, Set.Domain)
	if err := os.MkdirAll(Set.Directory, os.ModePerm); err != nil {
		return err
	}

	certificateOnly := bytes.TrimSpace(bytes.Replace(certificates.Certificate, certificates.IssuerCertificate, []byte{}, 1))
	if err := ioutil.WriteFile(Set.Certificate, certificateOnly, 0644); err != nil {
		logrus.Warnf(`[%s] failed to write certificate to archive folder.`, certificates.Domain)
		return err
	}

	if err := ioutil.WriteFile(Set.Chain, certificates.IssuerCertificate, 0644); err != nil {
		logrus.Warnf(`[%s] failed to write chain to archive folder.`, certificates.Domain)
		return err
	}

	fullChain := []byte(fmt.Sprintf("%s\n%s", certificateOnly, certificates.IssuerCertificate))
	if err := ioutil.WriteFile(Set.FullChain, fullChain, 0644); err != nil {
		logrus.Warnf(`[%s] failed to write full chain to archive folder.`, certificates.Domain)
		return err
	}

	if err := ioutil.WriteFile(Set.PrivateKey, certificates.PrivateKey, 0600); err != nil {
		logrus.Warnf(`[%s] failed to write private key to archive folder.`, certificates.Domain)
		return err
	}

	return nil
}

func (Set *FilePathSet) Activate(Live FilePathSet) error {
	logrus.Infof(`[INFO] [%s] files: Activating new live certificate files.`, Set.Domain)
	if err := os.MkdirAll(Live.Directory, os.ModePerm); err != nil {
		return err
	}

	Set.Certificate = strings.Replace(Set.Certificate, Set.Directory, `../../archive/`+Live.Domain, 1)
	Set.Chain = strings.Replace(Set.Chain, Set.Directory, `../../archive/`+Live.Domain, 1)
	Set.FullChain = strings.Replace(Set.FullChain, Set.Directory, `../../archive/`+Live.Domain, 1)
	Set.PrivateKey = strings.Replace(Set.PrivateKey, Set.Directory, `../../archive/`+Live.Domain, 1)

	if err := activateFile(`certificate`, Set.Certificate, Live.Certificate); err != nil {
		return err
	}
	if err := activateFile(`chain`, Set.Chain, Live.Chain); err != nil {
		return err
	}
	if err := activateFile(`fullchain`, Set.FullChain, Live.FullChain); err != nil {
		return err
	}
	if err := activateFile(`private key`, Set.PrivateKey, Live.PrivateKey); err != nil {
		return err
	}

	logrus.Infof(`[INFO] [%s] files: Activated new live certificate files.`, Set.Domain)
	return nil
}

func activateFile(Code string, OldPath, LivePath string) error {
	if err := os.Remove(LivePath); err != nil && !os.IsNotExist(err) {
		logrus.Warnf(`failed to remove old %s: %s: %s`, Code, LivePath, err)
		return err
	}
	if err := os.Symlink(OldPath, LivePath); err != nil {
		logrus.Warnf(`failed to activate %s: %s to %s: %s`, Code, OldPath, LivePath, err)
		return err
	}
	return nil
}

func WriteFile(Path string, data []byte, perm os.FileMode) error {
	Directory := filepath.Dir(Path)
	if err := os.MkdirAll(Directory, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(Path, data, perm)
}

func WriteFileJSON(Path string, v interface{}, perm os.FileMode) error {
	Buffer := bytes.Buffer{}
	Encoder := json.NewEncoder(&Buffer)
	Encoder.SetIndent(``, `  `)
	if err := Encoder.Encode(v); err != nil {
		return err
	}
	return WriteFile(Path, Buffer.Bytes(), perm)
}

func ReadFileJSON(Path string, dest interface{}) error {
	Contents, err := ioutil.ReadFile(Path)
	if err != nil {
		return err
	}
	return json.Unmarshal(Contents, dest)
}
