package store

import (
	"fmt"

	"go.uber.org/zap"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Jobs struct {
	client *kubernetes.Clientset
	log    *zap.SugaredLogger
}

func NewJobs(client *kubernetes.Clientset) *Jobs {
	s := &Jobs{
		client: client,
		log:    log,
	}

	return s
}

func (s *Jobs) GetName() string {
	return "jobs"
}

func (s *Jobs) GetObjects() ([]meta.Object, error) {
	list, err := s.client.
		BatchV1().
		Jobs(core.NamespaceAll).
		List(meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetch jobs / %w", err)
	}

	objects := []meta.Object{}
	for i := range list.Items {
		objects = append(objects, &list.Items[i])
	}

	return objects, nil
}

func (s *Jobs) GetListWatch() *cache.ListWatch {
	listWatch := cache.NewListWatchFromClient(
		s.client.BatchV1().RESTClient(),
		"jobs",
		core.NamespaceAll,
		fields.Nothing(),
	)

	return listWatch
}

func (s *Jobs) GetRuntimeObject() runtime.Object {
	return new(batch.Job)
}
