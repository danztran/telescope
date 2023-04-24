package promscope

import (
	"fmt"

	"github.com/danztran/telescope/pkg/mapnode"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/common/model"
)

const (
	ConnectionMetric string = "scope_connection"
	DurationMetric   string = "scope_duration_seconds"
)

// Connection is a label set of scope metrics
type Connection struct {
	Topology             string `json:"topology" mapstructure:"topology"`
	Source               string `json:"src" mapstructure:"src"`
	SourceNamespace      string `json:"src_ns" mapstructure:"src_ns"`
	Destination          string `json:"dest" mapstructure:"dest"`
	DestinationNamespace string `json:"dest_ns" mapstructure:"dest_ns"`
	DestinationPort      string `json:"dest_port" mapstructure:"dest_port"`
}

// convertToMapConnection convert labelset data to mapnode Connection
func convertToMapConnection(set model.LabelSet) (*mapnode.Connection, error) {
	conn := new(Connection)
	data := map[string]string{}
	for k, v := range set {
		data[string(k)] = string(v)
	}

	if err := mapstructure.Decode(data, conn); err != nil {
		return nil, fmt.Errorf("error decode connection: %+v / %w", set, err)
	}

	mapconn := mapnode.Connection{
		Source:               conn.Source,
		SourceNamespace:      conn.SourceNamespace,
		Destination:          conn.Destination,
		DestinationNamespace: conn.DestinationNamespace,
		DestinationPort:      conn.DestinationPort,
	}

	return &mapconn, nil
}
