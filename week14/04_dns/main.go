package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// ========================================
// Week 14, Lesson 4: DNS Operations
// ========================================
// DNS (Domain Name System) translates human-readable domain names
// to IP addresses. Go's net package provides built-in DNS resolution
// functions. This lesson covers all major DNS lookup operations.
//
// Usage:
//   go run main.go                    # Lookup google.com
//   go run main.go example.com        # Lookup custom domain
// ========================================

func main() {
	domain := "google.com"
	if len(os.Args) > 1 {
		domain = os.Args[1]
	}

	fmt.Println("========================================")
	fmt.Println("DNS Operations in Go")
	fmt.Println("========================================")
	fmt.Printf("Target domain: %s\n", domain)

	// ========================================
	// 1. net.LookupHost — Basic Name Resolution
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("1. net.LookupHost")
	fmt.Println("========================================")

	// LookupHost returns the host's IP addresses as strings.
	// It uses the local resolver (system DNS settings).
	fmt.Printf("\nLooking up hosts for %q...\n", domain)

	start := time.Now()
	addrs, err := net.LookupHost(domain)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Found %d address(es) in %v:\n", len(addrs), elapsed)
		for _, addr := range addrs {
			// Determine if IPv4 or IPv6
			ip := net.ParseIP(addr)
			version := "IPv4"
			if ip != nil && ip.To4() == nil {
				version = "IPv6"
			}
			fmt.Printf("    %s (%s)\n", addr, version)
		}
	}

	// ========================================
	// 2. net.LookupIP — Typed IP Resolution
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. net.LookupIP")
	fmt.Println("========================================")

	// LookupIP returns net.IP objects (more useful than strings).
	// You can filter by IPv4 or IPv6.
	fmt.Printf("\nLooking up IPs for %q...\n", domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  IPv4 addresses:")
		for _, ip := range ips {
			if ip.To4() != nil {
				fmt.Printf("    %s\n", ip)
				fmt.Printf("      IsLoopback: %v\n", ip.IsLoopback())
				fmt.Printf("      IsPrivate:  %v\n", ip.IsPrivate())
				fmt.Printf("      IsGlobal:   %v\n", ip.IsGlobalUnicast())
			}
		}

		fmt.Println("  IPv6 addresses:")
		for _, ip := range ips {
			if ip.To4() == nil {
				fmt.Printf("    %s\n", ip)
			}
		}
	}

	// ========================================
	// 3. net.LookupMX — Mail Exchange Records
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. net.LookupMX")
	fmt.Println("========================================")

	// MX records determine which mail servers receive email for a domain.
	// Lower priority numbers are preferred.
	fmt.Printf("\nLooking up MX records for %q...\n", domain)

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else if len(mxRecords) == 0 {
		fmt.Println("  No MX records found.")
	} else {
		fmt.Printf("  Found %d MX record(s):\n", len(mxRecords))
		for _, mx := range mxRecords {
			fmt.Printf("    Priority: %-5d Host: %s\n", mx.Pref, mx.Host)
		}
		fmt.Println("\n  Lower priority = higher preference (tried first)")
	}

	// ========================================
	// 4. net.LookupCNAME — Canonical Name
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. net.LookupCNAME")
	fmt.Println("========================================")

	// CNAME records are aliases that point to another domain name.
	// For example, www.example.com might CNAME to example.com.
	cnameTarget := "www." + domain
	fmt.Printf("\nLooking up CNAME for %q...\n", cnameTarget)

	cname, err := net.LookupCNAME(cnameTarget)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  CNAME: %s -> %s\n", cnameTarget, cname)
		if cname == cnameTarget+"." {
			fmt.Println("  (Same as queried — no CNAME redirect)")
		}
	}

	// Try some well-known CNAMEs
	fmt.Println("\nWell-known CNAME examples:")
	cnameExamples := []string{"www.github.com", "www.amazon.com"}
	for _, target := range cnameExamples {
		cname, err := net.LookupCNAME(target)
		if err != nil {
			fmt.Printf("  %s: error - %v\n", target, err)
		} else {
			fmt.Printf("  %s -> %s\n", target, cname)
		}
	}

	// ========================================
	// 5. net.LookupNS — Name Server Records
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. net.LookupNS")
	fmt.Println("========================================")

	// NS records identify the authoritative name servers for a domain.
	fmt.Printf("\nLooking up NS records for %q...\n", domain)

	nsRecords, err := net.LookupNS(domain)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Found %d name server(s):\n", len(nsRecords))
		for _, ns := range nsRecords {
			fmt.Printf("    %s\n", ns.Host)
		}
	}

	// ========================================
	// 6. net.LookupTXT — TXT Records
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. net.LookupTXT")
	fmt.Println("========================================")

	// TXT records contain arbitrary text, often used for:
	// - SPF (email sender verification)
	// - Domain verification (Google, etc.)
	// - DKIM keys
	fmt.Printf("\nLooking up TXT records for %q...\n", domain)

	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Found %d TXT record(s):\n", len(txtRecords))
		for i, txt := range txtRecords {
			// Truncate long records
			display := txt
			if len(display) > 80 {
				display = display[:80] + "..."
			}
			fmt.Printf("    [%d] %s\n", i+1, display)
		}
	}

	// ========================================
	// 7. net.LookupSRV — Service Records
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. net.LookupSRV")
	fmt.Println("========================================")

	// SRV records specify the host, port, priority, and weight for
	// a specific service. Used by LDAP, SIP, XMPP, etc.
	fmt.Println("\nLooking up SRV records...")

	// Try common SRV lookups
	srvQueries := []struct {
		service, proto, name string
	}{
		{"xmpp-server", "tcp", "gmail.com"},
		{"imaps", "tcp", "gmail.com"},
		{"sip", "tcp", domain},
	}

	for _, q := range srvQueries {
		cname, addrs, err := net.LookupSRV(q.service, q.proto, q.name)
		if err != nil {
			fmt.Printf("  _%s._%s.%s: %v\n", q.service, q.proto, q.name, err)
			continue
		}
		fmt.Printf("  _%s._%s.%s (CNAME: %s):\n", q.service, q.proto, q.name, cname)
		for _, addr := range addrs {
			fmt.Printf("    %s:%d (priority=%d, weight=%d)\n",
				addr.Target, addr.Port, addr.Priority, addr.Weight)
		}
	}

	// ========================================
	// 8. Reverse DNS Lookup
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Reverse DNS Lookup")
	fmt.Println("========================================")

	// Reverse lookup: IP address -> domain name
	// Uses PTR records in the DNS.
	fmt.Println("\nReverse DNS lookups:")

	reverseLookups := []string{"8.8.8.8", "1.1.1.1", "208.67.222.222"}

	// Also try the first IP we found for the target domain
	if len(addrs) > 0 {
		reverseLookups = append(reverseLookups, addrs[0])
	}

	for _, ip := range reverseLookups {
		names, err := net.LookupAddr(ip)
		if err != nil {
			fmt.Printf("  %s -> error: %v\n", ip, err)
		} else if len(names) == 0 {
			fmt.Printf("  %s -> (no PTR record)\n", ip)
		} else {
			fmt.Printf("  %s -> %s\n", ip, strings.Join(names, ", "))
		}
	}

	// ========================================
	// 9. Custom Resolver
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("9. Custom Resolver")
	fmt.Println("========================================")

	// net.Resolver allows you to customize DNS resolution.
	// You can specify a custom DNS server or force Go's pure-Go resolver.

	fmt.Println("\nUsing custom resolver:")

	// Force the pure-Go resolver (not the C library)
	goResolver := &net.Resolver{
		PreferGo: true, // Use Go's built-in DNS resolver
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	goAddrs, err := goResolver.LookupHost(ctx, domain)
	if err != nil {
		fmt.Printf("  Go resolver error: %v\n", err)
	} else {
		fmt.Printf("  Go resolver found %d addresses for %s:\n", len(goAddrs), domain)
		for _, addr := range goAddrs {
			fmt.Printf("    %s\n", addr)
		}
	}

	// Custom DNS server resolver
	customResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// Use Google's public DNS (8.8.8.8) instead of system default
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	customAddrs, err := customResolver.LookupHost(ctx2, domain)
	if err != nil {
		fmt.Printf("  Custom resolver (8.8.8.8) error: %v\n", err)
	} else {
		fmt.Printf("  Custom resolver (8.8.8.8) found %d addresses:\n", len(customAddrs))
		for _, addr := range customAddrs {
			fmt.Printf("    %s\n", addr)
		}
	}

	// ========================================
	// 10. DNS Concepts Reference
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("10. DNS Concepts Reference")
	fmt.Println("========================================")

	fmt.Print(`
  DNS Record Types:
    A      IPv4 address            example.com -> 93.184.216.34
    AAAA   IPv6 address            example.com -> 2606:2800:220:1:...
    CNAME  Canonical name (alias)  www.example.com -> example.com
    MX     Mail exchange           example.com -> mail.example.com
    NS     Name server             example.com -> ns1.example.com
    TXT    Text records            example.com -> "v=spf1 ..."
    SRV    Service locator         _sip._tcp.example.com -> sipserver:5060
    PTR    Reverse lookup          34.216.184.93 -> example.com
    SOA    Start of authority       Domain admin info, serial number

  DNS Resolution Process:
    1. Application calls net.LookupHost("example.com")
    2. Check local cache (/etc/hosts, OS DNS cache)
    3. Query recursive resolver (ISP DNS, 8.8.8.8, etc.)
    4. Resolver queries root nameserver -> .com NS -> example.com NS
    5. Authoritative nameserver returns the answer
    6. Response cached at each level (TTL-based)

  Go DNS Resolution:
    - By default, uses the OS resolver (cgo on most platforms)
    - Set GODEBUG=netdns=go to force pure-Go resolver
    - Set GODEBUG=netdns=cgo to force cgo resolver
    - net.Resolver.PreferGo=true uses pure-Go (no cgo needed)
`)

	fmt.Println("========================================")
	fmt.Println("DNS Operations lesson complete!")
	fmt.Println("========================================")
}
