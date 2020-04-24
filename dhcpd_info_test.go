package dhcpdleasesreader

import (
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	debug := true

	info, err := NewDhcpdInfo(get_file(), debug)
	if err != nil {
		t.Fatalf("Error creating the info object: %s", err)
	}

	err = info.Read()
	if err != nil {
		t.Fatalf("Error re-reading the info object: %s", err)
	}
}

func get_file() string {
	if os.Getenv("DHCPD_LEASES_FILE") != "" {
		return os.Getenv("DHCPD_LEASES_FILE")
	}
	return "/var/lib/dhcp/dhcpd.leases"
}
