package collector

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/danztran/telescope/pkg/kube"
	"github.com/danztran/telescope/pkg/promscope"
	"github.com/danztran/telescope/pkg/scope"
	"github.com/danztran/telescope/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultLogger = utils.MustGetLogger("collector")

type Deps struct {
	Log    *zap.SugaredLogger
	Kube   kube.Kube
	Scope  scope.Scope
	Config Config
}

type Config struct {
	TopologyID      string         `mapstructure:"topology_id"`
	SkipPatterns    []string       `mapstructure:"skip_patterns"`
	MaxNodeHandlers uint           `mapstructure:"max_node_handlers"`
	Metrics         Metrics        `mapstructure:"metrics"`
	ResetInterval   *time.Duration `mapstructure:"reset_interval"`
	CollectDuration *time.Duration `mapstructure:"collect_duration"`
}

type Metrics struct {
	Subsystem string `mapstructure:"subsystem"`
	Namespace string `mapstructure:"namespace"`
}

type Collector interface {
	Collect(ctx context.Context) error
	Reset() error
	RunCollectInterval(ctx context.Context)
	RunResetInterval(ctx context.Context)
}

type client struct {
	config         Config
	log            *zap.SugaredLogger
	scope          scope.Scope
	kube           kube.Kube
	metric         *prometheus.GaugeVec
	durationMetric *prometheus.HistogramVec
	nodeCache      *NodeCache
}

func MustNew(deps Deps) Collector {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Collector, error) {
	config := deps.Config

	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      promscope.ConnectionMetric,
		Subsystem: config.Metrics.Subsystem,
		Namespace: config.Metrics.Namespace,
	}, []string{"topology", "src", "src_ns", "dest", "dest_ns", "dest_port"})

	if err := prometheus.Register(metric); err != nil {
		return nil, err
	}

	durationMetric := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      promscope.DurationMetric,
		Subsystem: config.Metrics.Subsystem,
		Namespace: config.Metrics.Namespace,
		Buckets:   []float64{10, 20, 30, 60, 90, 120, 150, 180, 240, 270, 320, 360, 480, 540, 600, 1000},
	}, []string{"topology"})

	if err := prometheus.Register(durationMetric); err != nil {
		return nil, err
	}

	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	nodeCache := NewNodeCache(config.TopologyID, deps.Scope)

	instance := &client{
		config:         config,
		log:            deps.Log,
		scope:          deps.Scope,
		kube:           deps.Kube,
		metric:         metric,
		durationMetric: durationMetric,
		nodeCache:      nodeCache,
	}

	return instance, nil
}

func (c *client) Collect(ctx context.Context) error {
	defer c.nodeCache.Reset()

	topologyID := c.config.TopologyID
	ts := time.Now()
	defer func() {
		c.durationMetric.WithLabelValues(topologyID).Observe(time.Since(ts).Seconds())
	}()

	topology, err := c.scope.GetTopology(ctx, topologyID)
	if err != nil {
		return err
	}
	c.log.Infof("request topology found %d %s", len(topology.Nodes), topologyID)

	nodeChan := make(chan scope.NodeSummary, c.config.MaxNodeHandlers)
	go func() {
		defer close(nodeChan)
		for _, nodeSummary := range topology.Nodes {
			nodeChan <- nodeSummary
		}
	}()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for i := uint(0); i < c.config.MaxNodeHandlers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for nodeSummary := range nodeChan {
				if ctx.Err() != nil {
					break
				}
				err := c.ExposeNodeMetrics(ctx, nodeSummary)
				if err != nil {
					c.log.Error(err)
				}
			}
		}()
	}

	return nil
}

func (c *client) ExposeNodeMetrics(ctx context.Context, nodeSummary scope.NodeSummary) error {
	// get detail node (include connections info)
	srcNode, err := c.nodeCache.Get(ctx, nodeSummary.ID)
	if err != nil {
		if utils.IsErrNotFound(err) {
			c.log.Warnf("not found node: %s", nodeSummary.ID)
			return nil
		}
		return err
	}

	valid, err := c.IsValidLabels(*srcNode)
	if err != nil || !valid {
		return err
	}

	srcObject, err := c.GetRootObjectByNode(*srcNode)
	if err != nil {
		c.log.Warn(err)
		return nil
	}

	connections := getOutgoingConnections(*srcNode)
	if connections == nil {
		c.log.Warnf("not found connections: %s", nodeSummary.ID)
		return nil
	}

	for _, conn := range connections {
		destNode, err := c.nodeCache.Get(ctx, conn.NodeID)
		if err != nil {
			if utils.IsErrNotFound(err) {
				c.log.Warnf("not found node: %s", conn.NodeID)
				return nil
			}
			continue
		}

		valid, err := c.IsValidLabels(*destNode)
		if err != nil || !valid {
			continue
		}

		destObject, err := c.GetRootObjectByNode(*destNode)
		if err != nil {
			c.log.Warn(err)
			continue
		}

		// destPort := c.GetConnectionPort(conn)
		destPort := getConnectionPort(conn)
		ports, err := c.GetPodExposePorts(*destNode)
		if err != nil {
			c.log.Warn(err)
			continue
		}

		validPort := func() bool {
			if len(ports) == 0 {
				return true
			}
			for _, port := range ports {
				if destPort == port {
					return true
				}
			}
			return false
		}()

		if !validPort {
			c.log.Debugf(
				`ignored connection from "%s" to "%s": dest port "%s" not found in pod: %+v`,
				nodeSummary.Label, conn.Label, destPort, ports)
			continue
		}

		labels := prometheus.Labels{
			"topology":  c.config.TopologyID,
			"src":       srcObject.GetName(),
			"src_ns":    srcObject.GetNamespace(),
			"dest":      destObject.GetName(),
			"dest_ns":   destObject.GetNamespace(),
			"dest_port": destPort,
		}

		c.metric.With(labels)
		c.log.Infof("exposed metric %s: %v", promscope.ConnectionMetric, labels)
	}

	return nil
}

func (c *client) GetRootObjectByNode(node scope.APINode) (meta.Object, error) {
	podUID := getPodUID(node)
	if podUID == "" {
		return nil, fmt.Errorf(`not found pod uid: node_label="%s"`, node.Node.Label)
	}

	rootObject := c.kube.GetRootObject(podUID)
	if rootObject == nil {
		return nil, fmt.Errorf(`not found root object by uid="%s": node_label="%s"`, podUID, node.Node.Label)
	}

	return rootObject, nil
}

func (c *client) GetPodExposePorts(node scope.APINode) ([]string, error) {
	podUID := getPodUID(node)
	if podUID == "" {
		return nil, fmt.Errorf(`not found pod uid: node_label="%s"`, node.Node.Label)
	}

	pod, err := c.kube.GetPod(podUID)
	if pod == nil || err != nil {
		return nil, err
	}

	ports := make([]string, 0)
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			ports = append(ports, strconv.Itoa(int(port.ContainerPort)))
		}
	}

	return ports, nil
}

func (c *client) IsValidLabels(node scope.APINode) (bool, error) {
	for _, pattern := range c.config.SkipPatterns {
		matched, err := regexp.MatchString(pattern, node.Node.Label)
		if err != nil {
			return false, err
		}
		if matched {
			c.log.Debugf(`ignored node: skip_pattern="%s" label="%s"`, pattern, node.Node.Label)
			return false, nil
		}
	}

	return true, nil
}

func (c *client) Reset() error {
	c.metric.Reset()
	return nil
}

func (c *client) RunResetInterval(ctx context.Context) {
	if c.config.ResetInterval == nil {
		c.log.Info("disabled resetting interval")
		return
	}

	utils.RunStateless(ctx, *c.config.ResetInterval, func() {
		err := c.Reset()
		if err != nil {
			c.log.Error(err)
		}
	})
}

func (c *client) RunCollectInterval(ctx context.Context) {
	if c.config.CollectDuration == nil {
		c.log.Info("disabled collecting interval")
		return
	}

	utils.RunStateful(ctx, *c.config.CollectDuration, func() {
		topologyID := c.config.TopologyID
		defer utils.LogDuration()(c.log, "collecting topology %s", topologyID)

		c.log.Infof("start collecting topology: %s", topologyID)
		err := c.Collect(ctx)
		if err != nil {
			c.log.Error(err)
		}
	})
}
