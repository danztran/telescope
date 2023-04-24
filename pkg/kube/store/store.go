package store

import (
	"sync"
	"time"

	"github.com/danztran/telescope/pkg/utils"
	"go.uber.org/zap"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var log = utils.MustGetLogger("store")

type Deps struct {
	Log    *zap.SugaredLogger
	Client *kubernetes.Clientset
}

type Store struct {
	m         sync.Map
	resources []Resource
	log       *zap.SugaredLogger
	client    *kubernetes.Clientset
}

func New(deps Deps) (*Store, error) {
	if deps.Log == nil {
		deps.Log = log
	}
	client := deps.Client

	resources := []Resource{
		NewCronJobs(client),
		NewDaemonSets(client),
		NewDeployments(client),
		NewJobs(client),
		NewPods(client),
		NewReplicaSets(client),
	}

	s := &Store{
		m:         sync.Map{},
		log:       deps.Log,
		resources: resources,
	}

	s.watch()
	s.fetch()

	return s, nil
}

func (s *Store) fetch() error {
	for _, rs := range s.resources {
		objects, err := rs.GetObjects()
		if err != nil {
			return err
		}
		for i := range objects {
			s.Set(objects[i])
		}
	}

	return nil
}

func (s *Store) watch() {
	for _, rs := range s.resources {
		go func(rs Resource) {
			rsName := rs.GetName()

			wrapHandler := func(event string, ptr interface{}, handler func(meta.Object)) {
				object, ok := ptr.(meta.Object)
				if !ok {
					s.log.Warnf("%s %s - Invalid object: %+v", event, rsName, ptr)
					return
				}
				ns := object.GetNamespace()
				name := object.GetName()
				s.log.Debugf("%s %s: %s / %s", event, rsName, ns, name)
				handler(object)
			}

			handlerFuncs := cache.ResourceEventHandlerFuncs{
				AddFunc: func(ptr interface{}) {
					wrapHandler("ADD", ptr, s.Set)
				},
				UpdateFunc: func(old, ptr interface{}) {
					wrapHandler("UPDATE", ptr, s.Set)
				},
				DeleteFunc: func(ptr interface{}) {
					wrapHandler("DELETE", ptr, s.Remove)
				},
			}

			_, controller := cache.NewInformer(
				rs.GetListWatch(),
				rs.GetRuntimeObject(),
				time.Second*0,
				handlerFuncs,
			)

			stopCh := make(chan struct{})
			defer close(stopCh)
			controller.Run(stopCh)
			<-stopCh
		}(rs)
	}
}

func (s *Store) Get(uid types.UID) meta.Object {
	val, ok := s.m.Load(uid)
	if !ok {
		return nil
	}

	object, ok := val.(meta.Object)
	if !ok {
		return nil
	}

	return object
}

func (s *Store) Set(object meta.Object) {
	s.m.Store(object.GetUID(), object)
}

func (s *Store) Remove(object meta.Object) {
	s.m.Delete(object.GetUID())
}
