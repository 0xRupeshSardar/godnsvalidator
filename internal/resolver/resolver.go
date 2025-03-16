package resolver

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/0xRupeshSardar/godnsvalidator/internal/config"
	"github.com/0xRupeshSardar/godnsvalidator/pkg/utils"
)

type Baseline struct {
	GoodIP   string
	NXDomain bool
}

func GetBaseline(cfg *config.Config) *Baseline {
	baselineServers := []string{"1.1.1.1:53", "8.8.8.8:53"}
	
	for _, server := range baselineServers {
		ip, nx := checkBaselineServer(server, cfg)
		if ip != "" {
			return &Baseline{GoodIP: ip, NXDomain: nx}
		}
	}
	return nil
}

func checkBaselineServer(server string, cfg *config.Config) (string, bool) {
	// Check known domains
	for _, domain := range []string{"telegram.com", "bet365.com"} {
		ips, err := Resolve(context.Background(), domain, server, cfg.Timeout)
		if err != nil || len(ips) == 0 {
			return "", false
		}
	}

	// Check NXDOMAIN for random subdomains
	for _, domain := range []string{"facebook.com", "google.com"} {
		testDomain := utils.RandomString(10) + "." + domain
		_, err := Resolve(context.Background(), testDomain, server, cfg.Timeout)
		if err == nil {
			return "", false
		}
	}

	// Check root domain
	rootIPs, err := Resolve(context.Background(), cfg.RootDomain, server, cfg.Timeout)
	if err != nil || len(rootIPs) == 0 {
		return "", false
	}

	// Check NXDOMAIN for random subdomain
	nxDomain := utils.RandomString(10) + "." + cfg.RootDomain
	_, nxErr := Resolve(context.Background(), nxDomain, server, cfg.Timeout)

	return rootIPs[0], isNXDomain(nxErr)
}

func Resolve(ctx context.Context, domain, server string, timeout int) ([]string, error) {
	client := dns.Client{Timeout: time.Duration(timeout) * time.Second}
	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	r, _, err := client.ExchangeContext(ctx, &msg, server)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}

func isNXDomain(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "NXDOMAIN")
}