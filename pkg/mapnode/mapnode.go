// Package mapnode use MetricsClient as a module
// to get connection data and normalize to usable information.
package mapnode

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/danztran/telescope/pkg/utils"
	"go.uber.org/zap"
)

var defaultLogger = utils.MustGetLogger("mapnode")

type Deps struct {
	Log           *zap.SugaredLogger
	MetricsClient MetricsClient
	Config        Config
}

type Config struct {
	GetConnectionsSince time.Duration  `mapstructure:"get_connections_since"`
	UpdateInterval      *time.Duration `mapstructure:"update_interval"`
}

type Mapnode interface {
	UpdateData(ctx context.Context) error
	GetNode(name string) *Node
	GetAllNodes() map[string]Node
	GetLastUpdated() time.Time
	SinceLastUpdated() string
	RunUpdateInterval(ctx context.Context)
}

type mapnode struct {
	mx      sync.RWMutex
	config  Config
	log     *zap.SugaredLogger
	metrics MetricsClient

	nodes       map[string]Node
	lastUpdated time.Time
}

func MustNew(deps Deps) Mapnode {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Mapnode, error) {
	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	m := &mapnode{
		config:  deps.Config,
		log:     deps.Log,
		metrics: deps.MetricsClient,
		nodes:   make(map[string]Node),
	}

	err := m.UpdateData(context.Background())

	return m, err
}

func (m *mapnode) RunUpdateInterval(ctx context.Context) {
	if m.config.UpdateInterval == nil {
		m.log.Info("disabled updating interval")
		return
	}

	utils.RunStateless(ctx, *m.config.UpdateInterval, func() {
		err := m.UpdateData(ctx)
		if err != nil {
			m.log.Error(err)
		}
	})
}

// UpdateData get connections by MetricsClient
// and normalize to usable information.
func (m *mapnode) UpdateData(ctx context.Context) error {
	to := time.Now()
	start := to.Add(-m.config.GetConnectionsSince)
	connections, err := m.metrics.GetConnections(ctx, start, to)
	if err != nil {
		return err
	}

	// remove duplicated & normalize connection info
	mapConns := make(map[string]Connection)
	for _, conn := range connections {
		src := regexpNodeName.ReplaceAllString(conn.Source, "")
		dest := regexpNodeName.ReplaceAllString(conn.Destination, "")
		key := fmt.Sprintf("%s -> %s", src, dest)
		conn.Source = src
		conn.Destination = dest
		mapConns[key] = conn
	}

	// create node map
	nodes := make(map[string]Node)
	for _, conn := range mapConns {
		src := conn.Source
		srcNs := conn.SourceNamespace
		dest := conn.Destination
		destNs := conn.DestinationNamespace
		destPort := conn.DestinationPort

		if ip := net.ParseIP(dest); ip != nil {
			// skip ip address
			continue
		}

		// map source to nodes
		nodeSource, ok := nodes[src]
		if !ok {
			nodeSource = Node{
				Name:      src,
				Inbounds:  []Inbound{},
				Outbounds: []Outbound{},
			}
		}
		nodeSource.Outbounds = append(nodeSource.Outbounds, Outbound{
			Name:      dest,
			Namespace: destNs,
			Port:      destPort,
		})
		nodes[src] = nodeSource

		// map dest to nodes
		nodeDest, ok := nodes[dest]
		if !ok {
			nodeDest = Node{
				Name:      dest,
				Inbounds:  []Inbound{},
				Outbounds: []Outbound{},
			}
		}

		nodeDest.Inbounds = append(nodeDest.Inbounds, Inbound{
			Name:      src,
			Namespace: srcNs,
		})
		nodes[dest] = nodeDest
	}

	m.mx.Lock()
	defer m.mx.Unlock()
	m.nodes = nodes
	m.lastUpdated = time.Now()

	m.log.Debugf("mapped nodes length: %d", len(m.nodes))

	return nil
}

// GetNode get node's inbounds, outbounds information
func (m *mapnode) GetNode(name string) *Node {
	m.mx.RLock()
	defer m.mx.RUnlock()

	node, ok := m.nodes[name]
	if !ok {
		return nil
	}

	return &node
}

// GetAllNodes return a clone mapped nodes
func (m *mapnode) GetAllNodes() map[string]Node {
	m.mx.RLock()
	defer m.mx.RUnlock()

	nodes := make(map[string]Node, len(m.nodes))
	for k, v := range m.nodes {
		nodes[k] = v
	}

	return nodes
}

func (m *mapnode) GetLastUpdated() time.Time {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return m.lastUpdated
}

func (m *mapnode) SinceLastUpdated() string {
	return utils.SinceTime(m.GetLastUpdated(), time.Second)
}
