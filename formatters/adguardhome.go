package formatters

import (
	"fmt"
	"io"
	"lancache-dns-generator/cachedomains"
	"net/http"
	"strings"
)

func GetAdGuardHomeList(w http.ResponseWriter, r *http.Request) {
	rewrite_ip := r.URL.Query().Get("ip")
	if rewrite_ip == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "missing ip parameter")
		return
	}

	domains, err := cachedomains.GetAllDomainFiles()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprint(err))
		fmt.Println("error:", err)
		return
	}

	output := ""

	domain_lines := strings.Split(domains, "\n")
	for _, line := range domain_lines {
		if line == "" {
			continue
		}

		// comments just pass through unmodified
		if strings.HasPrefix(line, "#") {
			output += line + "\n"
			continue
		}

		// prefix is | for single domains and || for wildcards
		prefix := "|"
		domain := strings.TrimPrefix(line, "*.")
		if domain != line {
			prefix = "||"
		}

		// ignore duplicates
		if strings.Contains(output, fmt.Sprintf("%s%s^$dnsrewrite", prefix, domain)) {
			continue
		}

		// write rules
		output += fmt.Sprintf("%s%s^$dnsrewrite=%s\n", prefix, domain, rewrite_ip)
		output += fmt.Sprintf("%s%s^$dnstype=AAAA\n", prefix, domain)
	}

	io.WriteString(w, output)
}
