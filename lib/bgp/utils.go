package bgp

import (
	"strconv"
	"strings"
)

// Utility function used to convert a community string to uint32
func community_aton(community string) uint32 {
	var sum uint32
	var words []string
	var w0 int
	var w1 int

	words = strings.Split(community, ":")
	w0, _ = strconv.Atoi(words[0])
	w1, _ = strconv.Atoi(words[1])

	sum += uint32(w0) << 16
	sum += uint32(w1)

	return sum
}

// Utility function used to add a default prefixlen to a prefix if needed
func add_cidr_mask(addr string) string {
	if strings.Contains(addr, "/") {
		return addr
	}

	if strings.Contains(addr, ":") {
		return addr + "/128"
	} else {
		return addr + "/32"
	}
}
