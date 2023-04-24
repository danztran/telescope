package mapnode

import (
	"context"
	"regexp"
	"time"
)

var (
	regexpNodeName = regexp.MustCompile(`(.*\(|\)|-\d+$)`)
)

type Node struct {
	Name      string    `json:"name"`
	Outbounds Outbounds `json:"outbounds"`
	Inbounds  Inbounds  `json:"inbounds"`
}

type Inbounds []Inbound
type Outbounds []Outbound

type Outbound struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Port      string `json:"port"`
}

type Inbound struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Connection struct {
	Source               string `json:"source"`
	SourceNamespace      string `json:"source_namespace"`
	Destination          string `json:"destination"`
	DestinationNamespace string `json:"destination_namespace"`
	DestinationPort      string `json:"destination_port"`
}

type MetricsClient interface {
	GetConnections(ctx context.Context, start time.Time, end time.Time) ([]Connection, error)
}
