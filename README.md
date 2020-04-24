dhcpdleasesreader
==============

A very simple golang wrapper around the dhcpd.leases file (see man page). The wrapper can be created with two parameters:
`info, err := NewDhcpdInfo(file string, debug bool)`

Parameters:
- **file** - The path on the filesystem to the `dhcpd.leases` file. This is usually `/var/lib/dhcpd/dhcpd.leases`.
- **debug** - Whether to print verbose debugging information as the file is parsed

Return:
- **DhcpdInfo** structure - a structure with stats and the list of clients
- **err** - usually only occurs if there is an issue reading the `file` parameter


**Notes**
The `Read()` function can be called any time to refresh the list of clients and regenerate stats. At least, this will be only as expensive as a `stat()` check of the file because if it has not changed since last read, the function returns immediately.
