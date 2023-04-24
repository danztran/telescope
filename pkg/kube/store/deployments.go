package store

import (
	"fmt"

	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1beta2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Deployments struct {
	client *kubernetes.Clientset
	log    *zap.SugaredLogger
}

func NewDeployments(client *kubernetes.Clientset) *Deployments {
	s := &Deployments{
		client: client,
		log:    log,
	}

	return s
}

func (s *Deployments) GetName() string {
	return "deployments"
}

func (s *Deployments) GetObjects() ([]meta.Object, error) {
	list, err := s.client.
		AppsV1beta2().
		Deployments(core.NamespaceAll).
		List(meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetch deployments / %w", err)
	}

	objects := []meta.Object{}
	for i := range list.Items {
		objects = append(objects, &list.Items[i])
	}

	return objects, nil
}

func (s *Deployments) GetListWatch() *cache.ListWatch {
	listWatch := cache.NewListWatchFromClient(
		s.client.AppsV1beta2().RESTClient(),
		"deployments",
		core.NamespaceAll,
		fields.Nothing(),
	)

	return listWatch
}

func (s *Deployments) GetRuntimeObject() runtime.Object {
	return new(apps.Deployment)
}
