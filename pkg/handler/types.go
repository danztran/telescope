package handler

import "github.com/danztran/telescope/pkg/mapnode"

type GetNodeResponse struct {
	Node        mapnode.Node `json:"node"`
	LastUpdated string       `json:"last_updated"`
}

type GetAllNodesResponse struct {
	Nodes       map[string]mapnode.Node `json:"nodes"`
	LastUpdated string                  `json:"last_updated"`
}

type GetAllNodesOptions struct {
	ForceUpdate bool `json:"force_update" form:"force_update" query:"force_update"`
}
