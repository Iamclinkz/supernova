package discovery

import (
	"context"
	"fmt"
	"net/http"
	"supernova/pkg/constance"
	"supernova/pkg/util"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// 给外部使用，用于指定k8s服务发现的初始化内容
const (
	K8sMiddlewareNamespaceFieldName = "Namespace"

	//K8sRegisterConfigHealthcheckPortFieldName 用于指定k8s检查自己健康的端口
	K8sRegisterConfigHealthcheckPortFieldName = "HealthCheckPort"
)

// k8s自用，用于约定yaml中，部分label的名称。Scheduler会通过label的名称分析Executor原信息
const (
	k8sYamlLabelProtocFieldName = "ExecutorProtoc"
	k8sYamlLabelTagFieldName    = "ExecutorTag"
)

func NewK8sMiddlewareConfig(namespace string) MiddlewareConfig {
	return MiddlewareConfig{
		K8sMiddlewareNamespaceFieldName: namespace,
	}
}

func NewK8sRegisterConfig(healthCheckPort string) RegisterConfig {
	return RegisterConfig{
		K8sRegisterConfigHealthcheckPortFieldName: healthCheckPort,
	}
}

type K8sDiscoveryClient struct {
	clientSet                   *kubernetes.Clientset
	httpServer                  *http.Server
	namespace                   string
	serviceName2ServiceInstance map[string][]*ExecutorServiceInstance
	executorServiceInstance     []*ExecutorServiceInstance
	mutex                       sync.Mutex
	middlewareConfig            MiddlewareConfig
	registerConfig              RegisterConfig
}

func newK8sDiscoveryClient(middlewareConfig MiddlewareConfig, registerConfig RegisterConfig) (ExecutorDiscoveryClient, error) {
	if middlewareConfig[K8sMiddlewareNamespaceFieldName] == "" {
		panic("")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientSet: %v", err)
	}
	klog.Infof("k8s discovery client init success, namespace:%+v", middlewareConfig[K8sMiddlewareNamespaceFieldName])

	return &K8sDiscoveryClient{
		clientSet:                   clientSet,
		namespace:                   middlewareConfig[K8sMiddlewareNamespaceFieldName],
		middlewareConfig:            middlewareConfig,
		registerConfig:              registerConfig,
		serviceName2ServiceInstance: make(map[string][]*ExecutorServiceInstance),
	}, nil
}

// Register k8s的register启动service的时候就已经帮忙做了，这里启动一下http探针的handler即可
func (k *K8sDiscoveryClient) Register(instance *ExecutorServiceInstance) error {
	klog.Infof("start register, instance:%+v", instance)
	if k.registerConfig[K8sRegisterConfigHealthcheckPortFieldName] == "" {
		panic("")
	}
	healthCheckPort := k.registerConfig[K8sRegisterConfigHealthcheckPortFieldName]

	k.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%v", healthCheckPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/healthz":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			case "/readyz":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not found"))
				//todo 删掉！
				panic("")
			}
		}),
	}

	go func() {
		klog.Infof("try start HTTP server for health check at port:%v", healthCheckPort)
		if err := k.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("Error starting HTTP server for health check: %v", err)
		}
	}()

	return nil
}

func (k *K8sDiscoveryClient) DeRegister(instanceId string) error {
	if k.httpServer != nil {
		klog.Infof("try stop HTTP server for health check")
		if err := k.httpServer.Shutdown(context.Background()); err != nil {
			klog.Errorf("Error stopping HTTP server for health check: %v", err)
		}
	}

	return nil
}

func (k *K8sDiscoveryClient) DiscoverServices() []*ExecutorServiceInstance {
	klog.Tracef("start to DiscoverServices")
	if k.executorServiceInstance != nil {
		return k.executorServiceInstance
	}
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.executorServiceInstance != nil {
		return k.executorServiceInstance
	}

	k.executorServiceInstance = k.getInstancesFromServices()
	go func() {
		watcher, err := k.clientSet.CoreV1().Services(k.namespace).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: constance.K8sExecutorLabelName,
			Watch:         true,
		})
		if err != nil {
			klog.Errorf("DiscoverServices watch error: %v", err)
			return
		}
		defer watcher.Stop()

		for event := range watcher.ResultChan() {
			klog.Infof("discovery fetched event:%+v", event)

			svc, ok := event.Object.(*corev1.Service)
			if !ok {
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				var endpoints *corev1.Endpoints
				var err error

				// 重试获取 Endpoint
				for i := 0; i < 3; i++ {
					endpoints, err = k.clientSet.CoreV1().Endpoints(k.namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
					if err == nil {
						break
					}
					klog.Tracef("Failed to get endpoints for service %s: %v in namespace:%v, retrying...", svc.Name, err, k.namespace)
					time.Sleep(1 * time.Second)
				}

				if err != nil {
					klog.Errorf("Failed to get endpoints for service %s: %v in namespace:%v after retries", svc.Name, err, k.namespace)
					continue
				}

				instances := k.getInstancesFromEndpoints(endpoints, svc.Labels)
				k.serviceName2ServiceInstance[svc.Name] = instances
				k.refreshExecutorServiceInstance()
				klog.Infof("discovery refreshExecutorServiceInstance:%v", k.serviceName2ServiceInstance)
			case watch.Deleted:
				delete(k.serviceName2ServiceInstance, svc.Name)
				k.refreshExecutorServiceInstance()
				klog.Infof("discovery refreshExecutorServiceInstance:%v", k.serviceName2ServiceInstance)
			default:
				klog.Errorf("miss handle event type:%v", event.Type)
			}
		}
	}()

	return k.executorServiceInstance
}

func (k *K8sDiscoveryClient) getInstancesFromServices() []*ExecutorServiceInstance {
	services, err := k.clientSet.CoreV1().Services(k.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: constance.K8sExecutorLabelName,
	})
	if err != nil {
		klog.Errorf("Failed to list services: %v", err)
		return nil
	}

	var ret []*ExecutorServiceInstance
	for _, svc := range services.Items {
		endpoints, err := k.clientSet.CoreV1().Endpoints(k.namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get endpoints for service %s: %v", svc.Name, err)
			continue
		}

		instanceFromEndpoint := k.getInstancesFromEndpoints(endpoints, svc.Labels)
		ret = append(ret, instanceFromEndpoint...)
		k.serviceName2ServiceInstance[svc.Name] = instanceFromEndpoint
	}

	return ret
}

func (k *K8sDiscoveryClient) getInstancesFromEndpoints(endpoints *corev1.Endpoints, labels map[string]string) []*ExecutorServiceInstance {
	tags, protoc, err := extractTagsAndProtoc(labels)
	if err != nil {
		//todo 删掉
		panic(err)
	}

	var instances []*ExecutorServiceInstance
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			podName := address.TargetRef.Name
			if podName == "" {
				klog.Warnf("find podName empty:%+v", address)
				continue
			}

			instance := &ExecutorServiceInstance{
				InstanceId: podName,
				ExecutorServiceServeConf: ExecutorServiceServeConf{
					Protoc: protoc,
					Host:   address.IP,
					Port:   int(subset.Ports[0].Port),
				},
				Tags: tags,
			}
			instances = append(instances, instance)
		}
	}

	return instances
}

func extractTagsAndProtoc(labels map[string]string) ([]string, ProtocType, error) {
	var (
		tags   []string
		protoc ProtocType
	)

	protocStr, ok := labels[k8sYamlLabelProtocFieldName]
	if !ok {
		return nil, "", fmt.Errorf("can not find proto in labels:%v", labels)
	} else {
		switch ProtocType(protocStr) {
		case ProtocTypeGrpc:
			protoc = ProtocTypeGrpc
		case ProtocTypeHttp:
			protoc = ProtocTypeHttp
		default:
			return nil, "", fmt.Errorf("can not decode proto type:%v", protocStr)
		}
	}

	tagStr, ok := labels[k8sYamlLabelTagFieldName]
	if !ok || tagStr == "" {
		return nil, "", fmt.Errorf("can not find tags in labels:%v", labels)
	}
	tags = util.DecodeTags(tagStr)

	return tags, protoc, nil
}

func (k *K8sDiscoveryClient) refreshExecutorServiceInstance() {
	tmp := make([]*ExecutorServiceInstance, 0, len(k.serviceName2ServiceInstance))

	for _, instance := range k.serviceName2ServiceInstance {
		tmp = append(tmp, instance...)
	}

	k.executorServiceInstance = tmp
}

var _ ExecutorDiscoveryClient = (*K8sDiscoveryClient)(nil)
