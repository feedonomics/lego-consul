package types

type Renewal struct {
	Account       string       `json:"account,omitempty"`
	Authenticator string       `json:"authenticator,omitempty"`
	Server        string       `json:"server,omitempty"`
	Domain        string       `json:"domain,omitempty"`
	SANs          []string     `json:"sans,omitempty"`
	Paths         RenewalPaths `json:"paths,omitempty"`
}

type RenewalPaths struct {
	Certificate string `json:"certificate,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	FullChain   string `json:"full_chain,omitempty"`
	Chain       string `json:"chain,omitempty"`
}
