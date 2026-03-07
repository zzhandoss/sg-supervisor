package license

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
)

func ComputeFingerprint() (string, []string, error) {
	signals := []string{"os=" + runtime.GOOS, "arch=" + runtime.GOARCH}

	hostname, err := os.Hostname()
	if err != nil {
		return "", nil, err
	}
	signals = append(signals, "hostname="+hostname)

	if machineID, ok := readMachineID(); ok {
		signals = append(signals, "machine-id="+machineID)
	}

	macs := collectMACs()
	if len(macs) > 0 {
		signals = append(signals, "macs="+strings.Join(macs, ","))
	}

	sort.Strings(signals)
	sum := sha256.Sum256([]byte(strings.Join(signals, "|")))
	return hex.EncodeToString(sum[:]), signals, nil
}

func readMachineID() (string, bool) {
	candidates := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			value := strings.TrimSpace(string(data))
			if value != "" {
				return value, true
			}
		}
	}
	return "", false
}

func collectMACs() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	macs := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || len(iface.HardwareAddr) == 0 {
			continue
		}
		macs = append(macs, iface.HardwareAddr.String())
	}
	sort.Strings(macs)
	return macs
}
