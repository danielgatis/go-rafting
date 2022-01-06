package rafting

import (
	"context"
	"time"
)

func startDiscoveryLookup(ctx context.Context, n *Node) error {
	tick := time.Tick(n.discoveryLookupInterval)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick:
			addrs, err := n.discovery.Lookup()
			if err != nil {
				n.logger.Error(err)
				continue
			}

			peers := getPeers(addrs)
			addPeer(n, peers)
			remPeer(n, peers)
		}
	}
}

func startDiscoveryRegister(ctx context.Context, n *Node) error {
	return n.discovery.Register(ctx)
}
