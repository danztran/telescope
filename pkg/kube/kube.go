package kube

import (
	"fmt"
	"sync"

	"github.com/danztran/telescope/pkg/kube/store"
	"github.com/danztran/telescope/pkg/utils"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	defaultLogger = utils.MustGetLogger("scope")
	DefaultConfig = Config{
		Kubeconfig: DefaultConfigPath,
	}
)

type Deps struct {
	Log    *zap.SugaredLogger
	Config Config
}

type Config struct {
	Kubeconfig string
}

type Kube interface {
	GetRootObject(uid string) meta.Object
	GetPod(uid string) (*core.Pod, error)
}

type kube struct {
	sync.RWMutex
	config Config
	log    *zap.SugaredLogger
	store  *store.Store
}

func MustNew(deps Deps) Kube {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Kube, error) {
	config := deps.Config
	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	client, err := NewClient(config.Kubeconfig)
	if err != nil {
		return nil, err
	}

	deps.Log.Infof("initating kube store")

	store, err := store.New(store.Deps{
		Client: client,
	})
	if err != nil {
		return nil, err
	}

	k := &kube{
		RWMutex: sync.RWMutex{},
		config:  config,
		log:     deps.Log,
		store:   store,
	}

	return k, nil
}

func (k *kube) GetRootObject(uid string) meta.Object {
	var root meta.Object
	for {
		object := k.store.Get(types.UID(uid))
		if object == nil {
			break
		}
		root = object
		refs := root.GetOwnerReferences()
		if len(refs) == 0 {
			break
		}
		uid = string(refs[0].UID)
	}

	return root
}

func (k *kube) GetPod(uid string) (*core.Pod, error) {
	object := k.store.Get(types.UID(uid))
	if object == nil {
		return nil, nil
	}

	pod, ok := object.(*core.Pod)
	if !ok {
		return nil, fmt.Errorf("object found but not a pod")
	}

	clonePod := *pod

	return &clonePod, nil
}
