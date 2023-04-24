package scope

import (
	"time"
)

// APITopology is returned by the /api/topology/{name} handler.
type APITopology struct {
	Nodes NodeSummaries `json:"nodes"`
}

// APINode is returned by the /api/topology/{name}/{id} handler.
type APINode struct {
	Node Node `json:"node"`
}

// Node is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type Node struct {
	NodeSummary
	Controls    []ControlInstance    `json:"controls"`
	Children    []NodeSummaryGroup   `json:"children,omitempty"`
	Connections []ConnectionsSummary `json:"connections,omitempty"`
}

// NodeSummaryGroup is a topology-typed group of children for a Node.
type NodeSummaryGroup struct {
	ID         string        `json:"id"`
	Label      string        `json:"label"`
	Nodes      []NodeSummary `json:"nodes"`
	TopologyID string        `json:"topologyId"`
	Columns    []Column      `json:"columns"`
}

// ConnectionsSummary is the table of connection to/form a node
type ConnectionsSummary struct {
	ID          string       `json:"id"`
	TopologyID  string       `json:"topologyId"`
	Label       string       `json:"label"`
	Columns     []Column     `json:"columns"`
	Connections []Connection `json:"connections"`
}

// Connection is a row in the connections table.
type Connection struct {
	ID         string        `json:"id"`     // ID of this element in the UI.  Must be unique for a given ConnectionsSummary.
	NodeID     string        `json:"nodeId"` // ID of a node in the topology. Optional, must be set if linkable is true.
	Label      string        `json:"label"`
	LabelMinor string        `json:"labelMinor,omitempty"`
	Metadata   []MetadataRow `json:"metadata,omitempty"`
}

// ControlInstance contains a control description, and all the info
// needed to execute it.
type ControlInstance struct {
	ProbeID string
	NodeID  string
	Control Control
}

// Controls describe the control tags within the Nodes
type Controls map[string]Control

// A Control basically describes an RPC
type Control struct {
	ID           string `json:"id"`
	Human        string `json:"human"`
	Icon         string `json:"icon"` // from https://fortawesome.github.io/Font-Awesome/cheatsheet/ please
	Confirmation string `json:"confirmation,omitempty"`
	Rank         int    `json:"rank"`
}

// NodeSummaries is a set of NodeSummaries indexed by ID.
type NodeSummaries map[string]NodeSummary

// StringSet is a sorted set of unique strings. Clients must use the Add
// method to add strings.
type StringSet []string

type IDList StringSet

// BasicNodeSummary is basic summary information about a Node,
// sufficient for rendering links to the node.
type BasicNodeSummary struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	LabelMinor string `json:"labelMinor"`
	Rank       string `json:"rank"`
	Shape      string `json:"shape,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Stack      bool   `json:"stack,omitempty"`
	Pseudo     bool   `json:"pseudo,omitempty"`
}

// Parent is the information needed to build a link to the parent of a Node.
type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

// NodeSummary is summary information about a Node.
type NodeSummary struct {
	BasicNodeSummary
	Metadata  []MetadataRow `json:"metadata,omitempty"`
	Parents   []Parent      `json:"parents,omitempty"`
	Metrics   []MetricRow   `json:"metrics,omitempty"`
	Tables    []Table       `json:"tables,omitempty"`
	Adjacency IDList        `json:"adjacency,omitempty"`
}

// MetadataRow is a row for the metadata table.
type MetadataRow struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Value    string  `json:"value"`
	Priority float64 `json:"priority,omitempty"`
	Datatype string  `json:"dataType,omitempty"`
	Truncate int     `json:"truncate,omitempty"`
}

// MetricRow is a tuple of data used to render a metric as a sparkline and
// accoutrements.
type MetricRow struct {
	ID         string
	Label      string
	Format     string
	Group      string
	Value      float64
	ValueEmpty bool
	Priority   float64
	URL        string
	Metric     *Metric
}

// Metric is a list of timeseries data with some metadata. Clients must use the
// Add method to add values.  Metrics are immutable.
type Metric struct {
	Samples []Sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

// Sample is a single datapoint of a metric.
type Sample struct {
	Timestamp time.Time `json:"date"`
	Value     float64   `json:"value"`
}

// Table is the type for a table in the UI.
type Table struct {
	ID              string   `json:"id"`
	Label           string   `json:"label"`
	Type            string   `json:"type"`
	Columns         []Column `json:"columns"`
	Rows            []Row    `json:"rows"`
	TruncationCount int      `json:"truncationCount,omitempty"`
}

// Column provides special json serialization for column ids, so they include
// their label for the frontend.
type Column struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	DefaultSort bool   `json:"defaultSort"`
	Datatype    string `json:"dataType"`
}

// Row is the type that holds the table data for the UI. Entries map from column ID to cell value.
type Row struct {
	ID      string            `json:"id"`
	Entries map[string]string `json:"entries"`
}
