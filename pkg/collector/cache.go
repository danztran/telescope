package collector

import (
	"context"
	"sync"

	"github.com/danztran/telescope/pkg/scope"
	"go.uber.org/zap"
)

type NodeCache struct {
	m          sync.Map
	topologyID string
	scope      scope.Scope
	log        *zap.SugaredLogger
}

func NewNodeCache(topologyID string, scope scope.Scope) *NodeCache {
	c := NodeCache{
		m:          sync.Map{},
		topologyID: topologyID,
		scope:      scope,
		log:        defaultLogger,
	}

	return &c
}

func (c *NodeCache) Get(ctx context.Context, nodeID string) (*scope.APINode, error) {
	node := c.GetCache(nodeID)
	if node == nil {
		var err error
		node, err = c.scope.GetNode(ctx, c.topologyID, nodeID)
		if err != nil {
			return nil, err
		}

	}

	return node, nil
}

func (c *NodeCache) GetCache(nodeID string) *scope.APINode {
	val, ok := c.m.Load(nodeID)
	if !ok {
		return nil
	}

	node, ok := val.(scope.APINode)
	if !ok {
		return nil
	}

	return &node
}

func (c *NodeCache) Set(node scope.APINode) {
	c.m.Store(node.Node.ID, node)
}

func (c *NodeCache) Reset() {
	c.m.Range(func(key interface{}, _ interface{}) bool {
		c.m.Delete(key)
		return true
	})
}
