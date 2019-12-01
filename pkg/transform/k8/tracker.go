package k8

import (
	"path/filepath"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/pkg/logging"
	"go.uber.org/zap"
	fsnotify "gopkg.in/fsnotify/fsnotify.v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Resync period for the kube controller loop.
	resyncPeriod = 30 * time.Minute
)

type tracker interface {
	Get(string, string) *v1.Pod
}

type podTracker struct {
	client *kubernetes.Clientset

	// The name of the node that we are running on.
	NodeName string
	cache    *lru.Cache
}

func newTracker(conf Config) (tracker, error) {
	k8, err := newK8(conf.K8ConfigPath)
	if err != nil {
		return nil, err
	}
	tracker := newPodTracker(k8, conf.NodeName, conf.MaxPodsCache)
	tracker.watchForPods()
	return tracker, nil
}

func watchForK8ConfigFile(client *Client, conf Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	err = watcher.Add(filepath.Dir(conf.K8ConfigPath))
	if err != nil {
		panic(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			logging.Logger.Info("Got event for file", zap.String("name", event.Name))
			if event.Name == conf.K8ConfigPath {
				logging.Logger.Info("Creating new tracker for k8 config file")
				tracker, err := newTracker(conf)
				if err != nil {
					logging.Error(errors.Wrap(err, "Got error creating k8 tracker on fsnotify"))
				} else {
					client.tracker = tracker
					watcher.Close()
					return
				}
			}
		}
	}
}

func newK8(k8ConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", k8ConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "error getting config from path")
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "error constructing k8 client from config")
	}
	return clientset, nil
}

func newPodTracker(client *kubernetes.Clientset, nodeName string, maxPods int) *podTracker {
	cache, err := lru.New(maxPods)
	if err != nil {
		panic(err)
	}
	return &podTracker{
		NodeName: nodeName,
		cache:    cache,
		client:   client,
	}
}

func (t *podTracker) watchForPods() {
	err, podController := kcache.NewInformer(
		kcache.NewListWatchFromClient(t.client.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything()),
		&v1.Pod{},
		resyncPeriod,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    t.OnAdd,
			DeleteFunc: t.OnDelete,
			UpdateFunc: t.OnUpdate,
		},
	)
	if err != nil {
		panic(err)
	}
	go podController.Run(wait.NeverStop)
}

func (t *podTracker) Get(namespaceName, podName string) *v1.Pod {
	if val, ok := t.cache.Get(t.cacheKey(namespaceName, podName)); ok {
		return val.(*v1.Pod)
	}
	pod, err := t.client.CoreV1().Pods(namespaceName).Get(podName, metav1.GetOptions{})
	if err == nil {
		t.cache.ContainsOrAdd(t.cacheKey(namespaceName, podName), pod)
		return pod
	}
	return nil
}

func (t *podTracker) OnAdd(obj interface{}) {
	if pod, ok := obj.(*v1.Pod); ok {
		if t.canTrackPod(pod) {
			t.cache.Add(t.cacheKey(pod.Namespace, pod.Name), pod)
			logging.Logger.Info("Pod added", zap.String("namespace", pod.Namespace), zap.String("pod", pod.Name))
		}
	}
}

// Called with two api.Pod objects, with the first being the old version, and
// the second being the new version.
// It is invoked synchronously along with OnAdd and OnDelete.
func (t *podTracker) OnUpdate(oldObj, newObj interface{}) {
	_, ok1 := oldObj.(*v1.Pod)
	newPod, ok2 := newObj.(*v1.Pod)
	if !ok1 || !ok2 {
		return
	}
	if t.canTrackPod(newPod) {
		t.cache.Add(t.cacheKey(newPod.Namespace, newPod.Name), newPod)
		logging.Logger.Info("Pod updated", zap.String("namespace", newPod.Namespace), zap.String("pod", newPod.Name))
	}
}

// Called with an api.Pod object when the pod has been deleted.
// It is invoked synchronously along with OnAdd and OnUpdate.
func (t *podTracker) OnDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		deletedObj, dok := obj.(kcache.DeletedFinalStateUnknown)
		if dok {
			pod, ok = deletedObj.Obj.(*v1.Pod)
		}
	}
	if !ok {
		return
	}
	t.cache.Remove(t.cacheKey(pod.Namespace, pod.Name))
	logging.Logger.Info("Pod deleted", zap.String("namespace", pod.Namespace), zap.String("pod", pod.Name))
}

func (t *podTracker) cacheKey(namespaceName, podName string) string {
	return namespaceName + "_" + podName
}

// Returns true if the pod is schedule our node.
func (t *podTracker) canTrackPod(pod *v1.Pod) bool {
	if pod.Spec.NodeName == "" {
		return false
	} else if t.NodeName != "" && t.NodeName != pod.Spec.NodeName {
		return false
	}
	return true
}
