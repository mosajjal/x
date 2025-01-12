// Package ip contains the tools to set up a lookup system for a list of IPs and CIDR blocks.
// it can be presented as a trie, and it also has UDP, TCP, and HTTP servers.
package ip

import (
	"net"
	"sync"

	"github.com/yl2chen/cidranger"
)

// Looker is a wrapper around the cidranger.Ranger type
type Looker struct {
	ranger cidranger.Ranger
	mu     sync.RWMutex
}

// NewTrie creates a new trie ranger with the given IPnets
func NewTrie(IPnets ...*net.IPNet) *Looker {
	ranger := cidranger.NewPCTrieRanger()

	for _, IPnet := range IPnets {
		ranger.Insert(cidranger.NewBasicRangerEntry(*IPnet))
	}
	return &Looker{ranger: ranger}
}

// Insert adds the given CIDR block to the trie
// IMPORTANT NOTE: this function activates the lock
// which means it's not free.
func (r *Looker) Insert(IPnet *net.IPNet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ranger.Insert(cidranger.NewBasicRangerEntry(*IPnet))
}

// Contains checks if the given IP is in the trie
// Returns true if the IP is in the trie, false otherwise
func (r *Looker) Contains(ip net.IP) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	contains, _ := r.ranger.Contains(ip)
	return contains
}

// Reload replaces the current trie with a new one
func (r *Looker) Reload(IPnets ...*net.IPNet) {
	// Create new ranger outside of lock
	newRanger := cidranger.NewPCTrieRanger()
	for _, IPnet := range IPnets {
		newRanger.Insert(cidranger.NewBasicRangerEntry(*IPnet))
	}

	// Brief lock only for swapping
	r.mu.Lock()
	r.ranger = newRanger
	r.mu.Unlock()
}
