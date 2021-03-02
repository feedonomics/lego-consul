package utility

import "strings"

func ParseSANs(SANs []string) []string {
	var out []string
	for _, v := range SANs {
		parts := strings.Split(v, `,`)
		for _, part := range parts {
			part := strings.TrimSpace(part)
			if part != `` {
				out = append(out, part)
			}
		}
	}
	return out
}
