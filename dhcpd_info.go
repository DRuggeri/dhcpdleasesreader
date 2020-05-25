package dhcpdleasesreader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	mux sync.Mutex
)

type DhcpdInfo struct {
	file    string
	debug   bool
	ModTime time.Time
	Leases  map[string]*DhcpdLease
	Expired int
	Valid   int
}

type DhcpdLease struct {
	Hostname         string
	Starts           time.Time
	Ends             time.Time
	Cltt             time.Time
	Uid              string
	State            string
	Next             string
	Rewind           string
	Hardware_type    string
	Hardware_address string
	Ddns_fwd_name    string
	Ddns_rev_name    string
	Ddns_dhcid       string
	Identifier       string
}

func NewDhcpdInfo(i_file string, i_debug bool) (*DhcpdInfo, error) {
	info := DhcpdInfo{
		file:  i_file,
		debug: i_debug,
	}
	if info.debug {
		log.Printf("dhcpd_info.go: Constructing debug info for file %s\n", i_file)
	}

	err := info.Read()

	return &info, err
}

func (info *DhcpdInfo) Read() error {
	mux.Lock()
	defer mux.Unlock()

	fileInfo, err := os.Stat(info.file)
	if nil != err {
		return fmt.Errorf("dhcpd_info.go: stat %s: `%v`", info.file, err)
	}

	/* Short cut - don't read again if the file hasn't changed */
	if !fileInfo.ModTime().After(info.ModTime) {
		if info.debug {
			log.Printf("dhcpd_info.go: File has not changed since last read. Not reading again.\n")
		}
		return nil
	}

	if info.debug {
		log.Printf("dhcpd_info.go: Opening file %s\n", info.file)
	}

	/* Do this AFTER acquiring the mutex so anyone waiting on the mutex is immediately released */
	info.ModTime = fileInfo.ModTime()
	if info.debug {
		log.Printf("dhcpd_info.go: File has changed since last read (`%v` > `%v`). Reading and parsing file.\n", fileInfo.ModTime(), info.ModTime)
	}

	file, err := os.Open(info.file)
	if err != nil {
		return fmt.Errorf("dhcpd_info.go: open `%v`: `%v`", info.file, err)
	}
	defer file.Close()

	var ptr *DhcpdLease
	var client string
	now := time.Now()
	Valid := 0
	Expired := 0
	Leases := make(map[string]*DhcpdLease)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if info.debug {
			log.Printf("dhcpd_info.go: read: %v\n", line)
		}
		line = strings.TrimSpace(line)
		line = strings.TrimRight(line, ";\n")
		parts := strings.Split(line, " ")

		switch parts[0] {
		case "lease":
			client = parts[1]
			if info.debug {
				log.Printf("dhcpd_info.go: detected lease begin for `%v`\n", client)
			}
			lease := DhcpdLease{}
			Leases[client] = &lease
			ptr = &lease

		case "client-hostname":
			name := strings.Join(parts[1:], " ")
			name = strings.Trim(name, "\" ")
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed hostname: `%v`\n", name)
			}
			ptr.Hostname = name

		case "starts":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed start time: `%v`\n", t)
				}
				ptr.Starts = t
			}

		case "ends":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed end time: `%v`\n", t)
				}
				ptr.Ends = t
			}

		case "cltt":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed cltt time: `%v`\n", t)
				}
				ptr.Cltt = t
			}

		case "binding":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed state: `%v`\n", parts[2])
			}
			ptr.State = parts[2]

		case "next":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed next: `%v`\n", parts[2])
			}
			ptr.Next = parts[2]

		case "rewind":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed rewind: `%v`\n", parts[2])
			}
			ptr.Rewind = parts[2]

		case "hardware":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed hardware type/address: `%v`/`%v`\n", parts[1], parts[2])
			}
			ptr.Hardware_type = parts[1]
			ptr.Hardware_address = parts[2]

		case "uid":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed uid: `%v`\n", parts[1])
			}
			ptr.Uid = parts[1]

		case "set":
			parts2 := strings.Split(line, "=")
			value := strings.Trim(parts2[1], "\" ")
			switch parts[1] {
			case "ddns-fwd-name":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_fwd_name: `%v`\n", value)
				}
				ptr.Ddns_fwd_name = value

			case "ddns-rev-name":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_rev_name: `%v`\n", value)
				}
				ptr.Ddns_rev_name = value

			case "ddns-dhcid":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_dhcid: `%v`\n", value)
				}
				ptr.Ddns_dhcid = value

			case "vendor-class-identifier":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed identifier: `%v`\n", value)
				}
				ptr.Identifier = value

			default:
				if info.debug {
					log.Printf("dhcpd_info.go:   WARN - unmatched set '`%v`' = `%v`\n", parts[1], value)
				}
			}

		case "}":
		case "authoring-byte-order":
		case "#":
		case "":
			//Do nothing

		case "tstp":
		case "tsfp":
			//Ignore failover protocol

		default:
			if info.debug {
				log.Printf("dhcpd_info.go:   WARN - unmatched '`%v`'\n", parts[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("dhcpd_info.go: scanner read of `%v`: `%v`", info.file, err)
	}

	for _, lease := range Leases {
		if now.After(lease.Ends) {
			Expired++
		} else {
			Valid++
		}
	}

	/* Wait until the last second to replace data in the struct */
	info.Leases = Leases
	info.Valid = Valid
	info.Expired = Expired
	return nil
}
