package main

import (
	"flag"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/Users/rex_shen/.kube/config", "location of kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	// 初始化 k8s client
	clientset, _ := kubernetes.NewForConfig(config)

	// 初始化 informer factory
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour*12)

	// 創建速率限制列隊 RateLimitingQueue，入參為 string
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

	// 對 pod 進行監聽
	podInformer := informerFactory.Core().V1().Pods()
	informer := podInformer.Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { onAddPod(obj, queue) },
		DeleteFunc: func(obj interface{}) { onDeletePod(obj, queue) },
		UpdateFunc: func(obj interface{}, newObj interface{}) { onUpdatePod(newObj, queue) },
	})

	// Controller
	controller := NewController(queue, podInformer.Informer().GetIndexer(), informer)
	stopper := make(chan struct{})
	defer close(stopper)

	// 啟動 informer
	informerFactory.Start(stopper)
	informerFactory.WaitForCacheSync(stopper)

	// 處理隊列事件
	go func() {
		for {
			if !controller.processNextItem() {
				break
			}
		}
	}()

	<-stopper
}

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.TypedRateLimitingInterface[string]
	informer cache.Controller
}

func NewController(queue workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// 調用到 Done, 代表 queue 裡的元素都處理完了
	defer c.queue.Done(key)

	// 打印出處理的 key
	err := c.syncToStdout(key)
	c.handleErr(err, key)
	return true
}

func (c *Controller) syncToStdout(key string) error {
	// 通過 key 從 indexer 中獲取完整的對象
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		fmt.Printf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}
	if !exists {
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		pod := obj.(*corev1.Pod)
		fmt.Printf("Sync/Add/Update for Pod %s\n", pod.Name)
		// if pod.Name == "test-deployment" {
		// 	time.Sleep(2 * time.Second)
		// 	// 這邊用來測試確認下 key 會不會放回隊列
		// 	return fmt.Errorf("simulated error for deployment %s", deployment.Name)
		// }
	}
	return nil
}

func (c *Controller) handleErr(err error, key string) {
	if err == nil {
		c.queue.Forget(key)
		return
	}
	if c.queue.NumRequeues(key) < 5 {
		fmt.Printf("error syncing %q: %v", key, err)
		fmt.Printf("Retry %d for key %s\n", c.queue.NumRequeues(key), key)
		c.queue.AddRateLimited(key)
		return
	}
	c.queue.Forget(key)
	fmt.Printf("Dropping pod %q out of the queue: %v\n", key, err)
}

func onAddPod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	// 生成 key
	key, err := cache.MetaNamespaceKeyFunc(obj) // namespace/name
	if err == nil {
		queue.Add(key)
	}
}

func onDeletePod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		queue.Add(key)
	}
}

func onUpdatePod(newObj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err == nil {
		queue.Add(key)
	}
}
