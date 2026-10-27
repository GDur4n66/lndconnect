package main

import (
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/glendc/go-external-ip"
	"github.com/jessevdk/go-flags"
	"github.com/skip2/go-qrcode"
)

// getLocalIP returns the LAN address of this machine. It first asks the
// kernel which source address it would use for an outbound route (a UDP
// "connect" sends no packets), then falls back to scanning interfaces,
// preferring private addresses over public ones.
func getLocalIP() string {
	if conn, err := net.Dial("udp4", "192.0.2.1:9"); err == nil {
		addr, ok := conn.LocalAddr().(*net.UDPAddr)
		conn.Close()
		if ok && addr.IP.IsPrivate() {
			return addr.IP.String()
		}
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	fallback := ""
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil || ipnet.IP.IsLinkLocalUnicast() {
				continue
			}
			if ipnet.IP.IsPrivate() {
				return ipnet.IP.String()
			}
			if fallback == "" {
				fallback = ipnet.IP.String()
			}
		}
	}
	return fallback
}

func getPublicIP() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

func main() {
	loadedConfig, err := loadConfig()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			fmt.Println(err)
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := displayLink(loadedConfig); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func displayLink(loadedConfig *config) error {
	var err error

	// host
	ipString := ""
	if loadedConfig.LndConnect.Host != "" {
		ipString = loadedConfig.LndConnect.Host
	} else if loadedConfig.LndConnect.LocalIp {
		ipString = getLocalIP()
		if ipString == "" {
			return errors.New("could not determine local ip")
		}
	} else if loadedConfig.LndConnect.Localhost {
		ipString = "127.0.0.1"
	} else {
		ipString, err = getPublicIP()
		if err != nil {
			return err
		}
	}

	ipString = net.JoinHostPort(ipString, fmt.Sprint(loadedConfig.LndConnect.Port))

	u := url.URL{Scheme: "lndconnect", Host: ipString}
	q := u.Query()

	// cert
	if !loadedConfig.LndConnect.NoCert {
		certBytes, err := os.ReadFile(loadedConfig.TLSCertPath)
		if err != nil {
			return err
		}

		block, _ := pem.Decode(certBytes)
		if block == nil || block.Type != "CERTIFICATE" {
			return fmt.Errorf("%s: failed to decode PEM block "+
				"containing certificate", loadedConfig.TLSCertPath)
		}

		certificate := b64.RawURLEncoding.EncodeToString(block.Bytes)

		q.Add("cert", certificate)
	}

	// macaroon
	var macBytes []byte
	if loadedConfig.LndConnect.Invoice {
		macBytes, err = os.ReadFile(loadedConfig.InvoiceMacPath)
	} else if loadedConfig.LndConnect.Readonly {
		macBytes, err = os.ReadFile(loadedConfig.ReadMacPath)
	} else {
		macBytes, err = os.ReadFile(loadedConfig.AdminMacPath)
	}

	if err != nil {
		return err
	}

	macaroonB64 := b64.RawURLEncoding.EncodeToString(macBytes)

	q.Add("macaroon", macaroonB64)

	// custom query
	for _, s := range loadedConfig.LndConnect.Query {
		// Split on the first "=" only, so values may contain "=".
		queryParts := strings.SplitN(s, "=", 2)

		if len(queryParts) != 2 || queryParts[0] == "" {
			return fmt.Errorf("invalid query argument: %s", s)
		}

		q.Add(queryParts[0], queryParts[1])
	}

	u.RawQuery = q.Encode()

	// generate link / QR Code
	if loadedConfig.LndConnect.Url {
		fmt.Println(u.String())
	} else if loadedConfig.LndConnect.Image {
		err := qrcode.WriteFile(u.String(), qrcode.Medium, 512, "lndconnect-qr.png")
		if err != nil {
			return err
		}
		fmt.Println("Wrote QR Code to file \"lndconnect-qr.png\"")
	} else {
		obj := qrcodeTerminal.New()
		obj.Get(u.String()).Print()
		fmt.Println("\n⚠️  Press \"cmd + -\" a few times to see the full QR Code!\nIf that doesn't work run \"lndconnect -j\" to get a code you can copy paste into the app.")
	}

	return nil
}
