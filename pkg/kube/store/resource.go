package store

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type Resource interface {
	GetName() string
	GetObjects() ([]meta.Object, error)
	GetListWatch() *cache.ListWatch
	GetRuntimeObject() runtime.Object
}
