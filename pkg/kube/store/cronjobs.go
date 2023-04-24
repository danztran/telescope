package store

import (
	"fmt"

	"go.uber.org/zap"
	batch "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type CronJobs struct {
	client *kubernetes.Clientset
	log    *zap.SugaredLogger
}

func NewCronJobs(client *kubernetes.Clientset) *CronJobs {
	s := &CronJobs{
		client: client,
		log:    log,
	}

	return s
}

func (s *CronJobs) GetName() string {
	return "cronjobs"
}

func (s *CronJobs) GetObjects() ([]meta.Object, error) {
	list, err := s.client.
		BatchV1beta1().
		CronJobs(core.NamespaceAll).
		List(meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetch cronjobs / %w", err)
	}

	objects := []meta.Object{}
	for i := range list.Items {
		objects = append(objects, &list.Items[i])
	}

	return objects, nil
}

func (s *CronJobs) GetListWatch() *cache.ListWatch {
	listWatch := cache.NewListWatchFromClient(
		s.client.BatchV1beta1().RESTClient(),
		"cronjobs",
		core.NamespaceAll,
		fields.Nothing(),
	)

	return listWatch
}

func (s *CronJobs) GetRuntimeObject() runtime.Object {
	return new(batch.CronJob)
}
