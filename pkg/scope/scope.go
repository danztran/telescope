package scope

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danztran/telescope/pkg/httpclient"
	"github.com/danztran/telescope/pkg/utils"
	"go.uber.org/zap"
)

var defaultLogger = utils.MustGetLogger("scope")

const (
	InboundID  = "incoming-connections"
	OutboundID = "outgoing-connections"

	LabelDocker = "docker_label_"
	LabelPodUID = "label_io.kubernetes.pod.uid"
)

type Deps struct {
	Log    *zap.SugaredLogger
	Config Config
}

type Config struct {
	Address string
}

type Scope interface {
	GetTopology(ctx context.Context, topologyID string) (*APITopology, error)
	GetNode(ctx context.Context, topologyID string, nodeID string) (*APINode, error)
}

type scope struct {
	log    *zap.SugaredLogger
	client httpclient.Client
	config Config
}

func MustNew(deps Deps) Scope {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Scope, error) {
	config := deps.Config

	httpclient, err := httpclient.NewClient(httpclient.Config{
		Address: config.Address,
	})
	if err != nil {
		return nil, err
	}

	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	c := &scope{
		config: deps.Config,
		log:    deps.Log,
		client: httpclient,
	}

	return c, nil
}

func (s *scope) GetTopology(ctx context.Context, topologyID string) (*APITopology, error) {
	defer utils.LogDuration()(s.log, "GetTopology %s", topologyID)

	url := s.client.URL("/api/topology/:topology", map[string]string{
		"topology": topologyID,
	})

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error create new request / %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	apiTopology := new(APITopology)
	_, _, err = s.client.Do(ctx, req, apiTopology)
	if err != nil {
		return nil, fmt.Errorf("error get topology / %w", err)
	}

	return apiTopology, nil
}

func (s *scope) GetNode(ctx context.Context, topologyID string, nodeID string) (*APINode, error) {
	defer utils.LogDuration()(s.log, "GetNode %s/%s", topologyID, nodeID)

	url := s.client.URL("/api/topology/:topology/:nodeID", map[string]string{
		"topology": topologyID,
		"nodeID":   nodeID,
	})

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error create new request / %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	apiNode := new(APINode)
	_, _, err = s.client.Do(ctx, req, apiNode)
	if err != nil {
		return nil, fmt.Errorf("error get node / %w", err)
	}

	return apiNode, nil
}
