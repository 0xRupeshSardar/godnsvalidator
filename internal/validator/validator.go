package validator

import (
	"context"
	"net"
	"sync"

	"github.com/0xRupeshSardar/godnsvalidator/internal/config"
	"github.com/0xRupeshSardar/godnsvalidator/internal/resolver"
	"github.com/0xRupeshSardar/godnsvalidator/pkg/utils"
)

var (
	ValidServers []string
	mu           sync.Mutex
)

func ValidateServers(ctx context.Context, cfg *config.Config, baseline *resolver.Baseline) {
	targets := filterTargets(cfg)
	work := make(chan string, len(targets))
	var wg sync.WaitGroup

	for i := 0; i < cfg.Threads; i++ {
		wg.Add(1)
		go worker(ctx, &wg, work, cfg, baseline)
	}

	go func() {
		for _, t := range targets {
			select {
			case work <- t:
			case <-ctx.Done():
				// Return to stop the entire loop
				return
			}
		}
		close(work)
	}()
	
	wg.Wait()
}

func filterTargets(cfg *config.Config) []string {
	filtered := make([]string, 0)
	excl := make(map[string]struct{})
	
	for _, e := range cfg.Exclusions {
		excl[e] = struct{}{}
	}

	for _, t := range cfg.Targets {
		if _, exists := excl[t]; !exists && utils.IsValidIP(t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}


func worker(ctx context.Context, wg *sync.WaitGroup, work <-chan string,
    cfg *config.Config, baseline *resolver.Baseline) {
    defer wg.Done()

    for server := range work {
        select {
        case <-ctx.Done():
            return
        default:
            if validateServer(ctx, server, cfg, baseline) {
                mu.Lock()
                ValidServers = append(ValidServers, server)
                mu.Unlock()
            }
        }
    }
}

func validateServer(ctx context.Context, server string, cfg *config.Config, baseline *resolver.Baseline) bool {
    // Check NXDOMAIN for known domains
    for _, domain := range []string{"facebook.com", "google.com"} {
        testDomain := utils.RandomString(10) + "." + domain
        _, err := resolver.Resolve(ctx, testDomain, net.JoinHostPort(server, "53"), cfg.Timeout)
        if err == nil {
            return false
        }
    }

    // Check root domain resolution
    ips, err := resolver.Resolve(ctx, cfg.RootDomain, net.JoinHostPort(server, "53"), cfg.Timeout)
    if err != nil || len(ips) == 0 {
        return false
    }

    // Validate against baseline IP
    if baseline == nil || ips[0] != baseline.GoodIP {
        return false
    }

    // Check random subdomain NXDOMAIN
    testDomain := utils.RandomString(10) + "." + cfg.RootDomain
    _, nxErr := resolver.Resolve(ctx, testDomain, net.JoinHostPort(server, "53"), cfg.Timeout)
    if !resolver.IsNXDomain(nxErr) {
        return false
    }

    return true
}
