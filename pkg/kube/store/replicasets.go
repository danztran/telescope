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

type ReplicaSets struct {
	client *kubernetes.Clientset
	log    *zap.SugaredLogger
}

func NewReplicaSets(client *kubernetes.Clientset) *ReplicaSets {
	s := &ReplicaSets{
		client: client,
		log:    log,
	}

	return s
}

func (s *ReplicaSets) GetName() string {
	return "replicasets"
}

func (s *ReplicaSets) GetObjects() ([]meta.Object, error) {
	list, err := s.client.
		AppsV1beta2().
		ReplicaSets(core.NamespaceAll).
		List(meta.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error fetch replicasets / %w", err)
	}

	objects := []meta.Object{}
	for i := range list.Items {
		objects = append(objects, &list.Items[i])
	}

	return objects, nil
}

func (s *ReplicaSets) GetListWatch() *cache.ListWatch {
	listWatch := cache.NewListWatchFromClient(
		s.client.AppsV1beta2().RESTClient(),
		"replicasets",
		core.NamespaceAll,
		fields.Nothing(),
	)

	return listWatch
}

func (s *ReplicaSets) GetRuntimeObject() runtime.Object {
	return new(apps.ReplicaSet)
}
