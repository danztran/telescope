package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/danztran/telescope/config"
	"github.com/danztran/telescope/pkg/collector"
	"github.com/danztran/telescope/pkg/handler"
	"github.com/danztran/telescope/pkg/kube"
	"github.com/danztran/telescope/pkg/mapnode"
	"github.com/danztran/telescope/pkg/promscope"
	"github.com/danztran/telescope/pkg/scope"
	"github.com/danztran/telescope/pkg/server"
	"github.com/danztran/telescope/pkg/utils"
	"github.com/spf13/cobra"
)

var log = utils.MustGetLogger("cmd")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "telescope",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ScopeClient := scope.MustNew(scope.Deps{
			Config: config.Values.Scope,
		})

		Kube := kube.MustNew(kube.Deps{
			Config: kube.DefaultConfig,
		})

		Collector := collector.MustNew(collector.Deps{
			Scope:  ScopeClient,
			Kube:   Kube,
			Config: config.Values.Collector,
		})

		Promscope := promscope.MustNew(promscope.Deps{
			Config: config.Values.Promscope,
		})

		Mapnode := mapnode.MustNew(mapnode.Deps{
			MetricsClient: Promscope,
			Config:        config.Values.Mapnode,
		})

		Handler := handler.MustNew(handler.Deps{
			Mapnode: Mapnode,
		})

		Server := server.MustNew(server.Deps{
			Handler: Handler,
			Config:  config.Values.Server,
		})

		wg := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())

		var err error
		go utils.RunJobsWithContext(
			ctx,
			&wg,
			Collector.RunCollectInterval,
			Collector.RunResetInterval,
			Mapnode.RunUpdateInterval,
			func(ctx context.Context) {
				err = Server.Run(ctx)
				if err != nil {
					err = fmt.Errorf("server error / %w", err)
				}
			},
		)

		utils.WaitToStop()
		log.Infof("terminating...")
		cancel()
		wg.Wait()

		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var err error
	log.Debugf("%+v", config.Values)

	timeStart := time.Now()
	err = rootCmd.Execute()
	execTime := time.Since(timeStart)
	log.Infof("execution time: %v", execTime)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
