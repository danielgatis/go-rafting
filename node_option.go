package rafting

import (
	"fmt"
	"time"

	"github.com/danielgatis/go-discovery"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type NodeOption func(*Node)

func WithLogger(logger logrus.FieldLogger) NodeOption {
	return func(n *Node) {
		n.logger = logger
	}
}

func WithDataDir(dataDir string) NodeOption {
	return func(n *Node) {
		n.dataDir = dataDir
	}
}

func WithStaticDiscovery(peers []string) NodeOption {
	return func(n *Node) {
		n.discovery = discovery.NewDummyDiscovery(peers, n.logger)
	}
}

func WithMdnsDiscovery() NodeOption {
	return func(n *Node) {
		n.discovery = discovery.NewMdnsDiscovery(fmt.Sprintf("raft:%s", n.id), "_raft._tcp", "local.", n.port, n.logger)
	}
}

func WithK8sDiscovery(clientset kubernetes.Interface, namespace string, portName string, labels map[string]string) NodeOption {
	return func(n *Node) {
		n.discovery = discovery.NewK8sDiscovery(clientset, namespace, portName, labels, n.logger)
	}
}

func WithDiscovery(discovery discovery.Discovery) NodeOption {
	return func(n *Node) {
		n.discovery = discovery
	}
}

func WithDiscoveryLookupInterval(interval time.Duration) NodeOption {
	return func(n *Node) {
		n.discoveryLookupInterval = interval
	}
}

func WithSerializer(marshalFn func(interface{}) ([]byte, error), unmarshalFn func([]byte, interface{}) error) NodeOption {
	return func(n *Node) {
		n.marshalFn = marshalFn
		n.unmarshalFn = unmarshalFn
	}
}
