package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/danztran/telescope/pkg/httpclient"
	"github.com/danztran/telescope/pkg/mapnode"
	"github.com/danztran/telescope/pkg/utils"
	"go.uber.org/zap"
)

var defaultLogger = utils.MustGetLogger("handler")

type Deps struct {
	Log     *zap.SugaredLogger
	Mapnode mapnode.Mapnode
}

type Handler interface {
	GetConnectionsByName(ctx context.Context, name string) (*GetNodeResponse, error)
	GetAllConnections(ctx context.Context, opt GetAllNodesOptions) (*GetAllNodesResponse, error)
}

type handler struct {
	log     *zap.SugaredLogger
	mapnode mapnode.Mapnode
}

func MustNew(deps Deps) Handler {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Handler, error) {
	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	h := &handler{
		log:     deps.Log,
		mapnode: deps.Mapnode,
	}
	return h, nil
}

func (h *handler) GetConnectionsByName(ctx context.Context, name string) (*GetNodeResponse, error) {
	node := h.mapnode.GetNode(name)
	if node == nil {
		return nil, &httpclient.ErrNotFound{
			Message: fmt.Sprintf("not found any node with name: %s", name),
		}
	}

	resp := &GetNodeResponse{
		Node:        *node,
		LastUpdated: h.mapnode.SinceLastUpdated(),
	}

	return resp, nil
}

func (h *handler) GetAllConnections(ctx context.Context, opt GetAllNodesOptions) (*GetAllNodesResponse, error) {
	nodes := h.mapnode.GetAllNodes()
	lastUpdated := h.mapnode.GetLastUpdated()

	if opt.ForceUpdate && time.Since(lastUpdated) > 60*time.Second {
		err := h.mapnode.UpdateData(ctx)
		if err != nil {
			return nil, fmt.Errorf("error update data / %w", err)
		}
	}

	resp := &GetAllNodesResponse{
		Nodes:       nodes,
		LastUpdated: h.mapnode.SinceLastUpdated(),
	}

	return resp, nil
}
