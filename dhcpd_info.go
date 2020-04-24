package dhcpdleasesreader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type DhcpdInfo struct {
	file    string
	debug   bool
	modTime time.Time
	Leases  map[string]*DhcpdLease
	Expired int
	Valid   int
}

type DhcpdLease struct {
	hostname         string
	starts           time.Time
	ends             time.Time
	cltt             time.Time
	uid              string
	state            string
	next             string
	rewind           string
	hardware_type    string
	hardware_address string
	ddns_fwd_name    string
	ddns_rev_name    string
	ddns_dhcid       string
	identifier       string
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
	if info.debug {
		log.Printf("dhcpd_info.go: Opening file %s\n", info.file)
	}

	fileInfo, err := os.Stat(info.file)
	if nil != err {
		return fmt.Errorf("dhcpd_info.go: stat %s: `%v`", info.file, err)
	}

	/* Short cut - don't read again if the file hasn't changed */
	if !fileInfo.ModTime().After(info.modTime) {
		if info.debug {
			log.Printf("dhcpd_info.go: File has not changed since last read. Not reading again.\n")
		}
		return nil
	}

	info.modTime = fileInfo.ModTime()
	if info.debug {
		log.Printf("dhcpd_info.go: File has changed since last read (`%v` > `%v`). Reading and parsing file.\n", fileInfo.ModTime(), info.modTime)
	}

	file, err := os.Open(info.file)
	if err != nil {
		return fmt.Errorf("dhcpd_info.go: open `%v`: `%v`", info.file, err)
	}
	defer file.Close()

	var ptr *DhcpdLease
	var client string
	now := time.Now()
	info.Valid = 0
	info.expired = 0
	info.Leases = make(map[string]*DhcpdLease)
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
			info.Leases[client] = &lease
			ptr = &lease

		case "client-hostname":
			name := strings.Join(parts[1:], " ")
			name = strings.Trim(name, "\" ")
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed hostname: `%v`\n", name)
			}
			ptr.hostname = name

		case "starts":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed start time: `%v`\n", t)
				}
				ptr.starts = t
			}

		case "ends":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed end time: `%v`\n", t)
				}
				ptr.ends = t
			}
			if now.After(t) {
				info.Expired++
			} else {
				info.Valid++
			}

		case "cltt":
			t, err := time.Parse(`2006/01/02 15:04:05`, fmt.Sprintf("%s %s", parts[2], parts[3]))
			if err != nil {
				log.Printf("Error parsing time: `%v`\n", err)
			} else {
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed cltt time: `%v`\n", t)
				}
				ptr.cltt = t
			}

		case "binding":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed state: `%v`\n", parts[2])
			}
			ptr.state = parts[2]

		case "next":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed next: `%v`\n", parts[2])
			}
			ptr.next = parts[2]

		case "rewind":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed rewind: `%v`\n", parts[2])
			}
			ptr.rewind = parts[2]

		case "hardware":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed hardware type/address: `%v`/`%v`\n", parts[1], parts[2])
			}
			ptr.hardware_type = parts[1]
			ptr.hardware_address = parts[1]

		case "uid":
			if info.debug {
				log.Printf("dhcpd_info.go:   Parsed uid: `%v`\n", parts[1])
			}
			ptr.uid = parts[1]

		case "set":
			parts2 := strings.Split(line, "=")
			value := strings.Trim(parts2[1], "\" ")
			switch parts[1] {
			case "ddns-fwd-name":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_fwd_name: `%v`\n", value)
				}
				ptr.ddns_fwd_name = value

			case "ddns-rev-name":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_rev_name: `%v`\n", value)
				}
				ptr.ddns_rev_name = value

			case "ddns-dhcid":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed ddns_dhcid: `%v`\n", value)
				}
				ptr.ddns_dhcid = value

			case "vendor-class-identifier":
				if info.debug {
					log.Printf("dhcpd_info.go:   Parsed identifier: `%v`\n", value)
				}
				ptr.identifier = value

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

	return nil
}
