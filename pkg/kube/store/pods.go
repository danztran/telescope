package store

import (
	"fmt"

	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Pods struct {
	client *kubernetes.Clientset
	log    *zap.SugaredLogger
}

func NewPods(client *kubernetes.Clientset) *Pods {
	s := &Pods{
		client: client,
		log:    log,
	}

	return s
}

func (s *Pods) GetName() string {
	return "pods"
}

func (s *Pods) GetObjects() ([]meta.Object, error) {
	list, err := s.client.
		CoreV1().
		Pods(core.NamespaceAll).
		List(meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetch pods / %w", err)
	}

	objects := []meta.Object{}
	for i := range list.Items {
		objects = append(objects, &list.Items[i])
	}

	return objects, nil
}

func (s *Pods) GetListWatch() *cache.ListWatch {
	listWatch := cache.NewListWatchFromClient(
		s.client.CoreV1().RESTClient(),
		"pods",
		core.NamespaceAll,
		fields.Nothing(),
	)

	return listWatch
}

func (s *Pods) GetRuntimeObject() runtime.Object {
	return new(core.Pod)
}
