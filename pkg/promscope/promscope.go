// Package promscope is a MetricsClient module get metrics from Prometheus
// and parse to mapnode models.
package promscope

import (
	"context"
	"fmt"
	"time"

	"github.com/danztran/telescope/pkg/mapnode"
	"github.com/danztran/telescope/pkg/utils"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

var defaultLogger = utils.MustGetLogger("promscope")

type Deps struct {
	Log    *zap.SugaredLogger
	Config Config
}

type Config struct {
	Prometheus         Prometheus    `mapstructure:"prometheus"`
	GetConnectionsStep time.Duration `mapstructure:"get_connections_step"`
}

type Prometheus struct {
	Address string  `mapstructure:"address"`
	Token   *string `mapstructure:"token"`
}

type Promscope interface {
	GetConnections(ctx context.Context, start time.Time, end time.Time) ([]mapnode.Connection, error)
}

type promscope struct {
	config  Config
	log     *zap.SugaredLogger
	promAPI promv1.API
}

func MustNew(deps Deps) Promscope {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

// New create new promscope interface
func New(deps Deps) (Promscope, error) {
	config := deps.Config

	promAPI := promv1.NewAPI(nil)

	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	p := &promscope{
		config:  config,
		log:     deps.Log,
		promAPI: promAPI,
	}

	return p, nil
}

// GetConnections get scope connections metrics from prometheus
// and parse to Connection model
func (p *promscope) GetConnections(ctx context.Context, start time.Time, end time.Time) ([]mapnode.Connection, error) {
	defer utils.LogDuration()(p.log, "GetConnections with [start:%v] [end:%v]", start, end)

	query := fmt.Sprintf("sum (%s) by (src, dest, dest_port, topology)", ConnectionMetric)
	val, warns, err := p.promAPI.QueryRange(ctx, query, promv1.Range{
		Start: start,
		End:   end,
		Step:  p.config.GetConnectionsStep,
	})
	if err != nil {
		return nil, fmt.Errorf("error query series: %s / %w", query, err)
	}
	if warns != nil {
		p.log.Warnf("warning get series: %s / %v", query, warns)
	}

	matrix, ok := val.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("invalid matrix / %s", val)
	}

	connections := make([]mapnode.Connection, len(matrix))
	p.log.Debugf("found %d matrix streams", len(connections))
	for i, v := range matrix {
		lbSet := model.LabelSet(v.Metric)
		conn, err := convertToMapConnection(lbSet)
		if err != nil {
			return nil, fmt.Errorf("error convert to map connection / %w", err)
		}
		connections[i] = *conn
	}

	return connections, nil
}
