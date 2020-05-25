package dhcpdleasesreader

import (
	"log"
	"os"
	"testing"
	"time"
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

func TestMux(t *testing.T) {
	debug := false

	info, err := NewDhcpdInfo(get_file(), debug)
	if err != nil {
		t.Fatalf("Error creating the info object: %s", err)
	}

	fxn := func(t *testing.T, info *DhcpdInfo, bg bool) {
		where := "foreground"
		if bg {
			where = "background"
		}
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			log.Printf("Attempting %s read...\n", where)
			err = info.Read()
			if err != nil {
				t.Fatalf("Error re-reading the info object: %s", err)
			}
		}
	}
	go fxn(t, info, true)
	fxn(t, info, false)
}

func get_file() string {
	if os.Getenv("DHCPD_LEASES_FILE") != "" {
		return os.Getenv("DHCPD_LEASES_FILE")
	}
	return "/var/lib/dhcp/dhcpd.leases"
}
