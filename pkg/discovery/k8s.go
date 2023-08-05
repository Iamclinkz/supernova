package discovery

//
//import (
//	"context"
//	"fmt"
//	"net/http"
//	"sync"
//	"time"
//
//	"github.com/cloudwego/kitex/pkg/klog"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/watch"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/rest"
//)
//
//// 给外部使用，用于指定k8s服务发现的初始化内容
//const (
//	K8sMiddlewareNamespaceFieldName = "Namespace"
//
//	//K8sRegisterConfigHealthcheckPortFieldName 用于指定k8s检查自己健康的端口
//	K8sRegisterConfigHealthcheckPortFieldName = "HealthCheckPort"
//)
//
//// k8s自用，用于约定yaml中，部分label的字段名称。Scheduler会通过label的名称解析对应的原信息
//const (
//	k8sYamlLabelProtocFieldName      = "SupernovaServiceProtoc"
//	k8sYamlLabelExtraConfigFieldName = "SupernovaServiceExtraConfig"
//)
//
//func NewK8sMiddlewareConfig(namespace string) MiddlewareConfig {
//	return MiddlewareConfig{
//		K8sMiddlewareNamespaceFieldName: namespace,
//	}
//}
//
//func NewK8sRegisterConfig(healthCheckPort string) RegisterConfig {
//	return RegisterConfig{
//		K8sRegisterConfigHealthcheckPortFieldName: healthCheckPort,
//	}
//}
//
//type K8sDiscoveryClient struct {
//	clientSet                           *kubernetes.Clientset
//	httpServer                          *http.Server
//	namespace                           string
//	k8sLabel2K8sService2ServiceInstance map[string]map[string][]*ServiceInstance //k8s的label到k8s的serviceName到ServiceInstance，go程不安全，要加锁CRUD
//	instancesMap                        sync.Map                                 //map[InstanceServiceName][]*ServiceInstance，注意instanceServiceName不等于k8sServiceName！
//	mutex                               sync.Mutex
//	middlewareConfig                    MiddlewareConfig
//	registerConfig                      RegisterConfig
//}
//
//func newK8sDiscoveryClient(middlewareConfig MiddlewareConfig, registerConfig RegisterConfig) (DiscoverClient, error) {
//	if middlewareConfig[K8sMiddlewareNamespaceFieldName] == "" {
//		panic("")
//	}
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
//	}
//
//	clientSet, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create clientSet: %v", err)
//	}
//	klog.Infof("k8s discovery client init success, namespace:%+v", middlewareConfig[K8sMiddlewareNamespaceFieldName])
//
//	return &K8sDiscoveryClient{
//		clientSet:                           clientSet,
//		namespace:                           middlewareConfig[K8sMiddlewareNamespaceFieldName],
//		middlewareConfig:                    middlewareConfig,
//		registerConfig:                      registerConfig,
//		k8sLabel2K8sService2ServiceInstance: make(map[string]map[string][]*ServiceInstance),
//	}, nil
//}
//
//// Register k8s的register启动service的时候就已经帮忙做了，这里启动一下http探针的handler即可
//func (k *K8sDiscoveryClient) Register(instance *ServiceInstance) error {
//	klog.Infof("start register, instance:%+v", instance)
//	if k.registerConfig[K8sRegisterConfigHealthcheckPortFieldName] == "" {
//		panic("")
//	}
//	healthCheckPort := k.registerConfig[K8sRegisterConfigHealthcheckPortFieldName]
//
//	k.httpServer = &http.Server{
//		Addr: fmt.Sprintf(":%v", healthCheckPort),
//		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			switch r.URL.Path {
//			case "/healthz":
//				w.WriteHeader(http.StatusOK)
//				_, _ = w.Write([]byte("OK"))
//			case "/readyz":
//				w.WriteHeader(http.StatusOK)
//				_, _ = w.Write([]byte("OK"))
//			default:
//				w.WriteHeader(http.StatusNotFound)
//				_, _ = w.Write([]byte("Not found"))
//				//todo 删掉！
//				panic("")
//			}
//		}),
//	}
//
//	go func() {
//		klog.Infof("try start HTTP server for health check at port:%v", healthCheckPort)
//		if err := k.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//			klog.Errorf("Error starting HTTP server for health check: %v", err)
//		}
//	}()
//
//	return nil
//}
//
//func (k *K8sDiscoveryClient) DeRegister(instanceId string) error {
//	if k.httpServer != nil {
//		klog.Infof("try stop HTTP server for health check")
//		if err := k.httpServer.Shutdown(context.Background()); err != nil {
//			klog.Errorf("Error stopping HTTP server for health check: %v", err)
//		}
//	}
//
//	return nil
//}
//
//// DiscoverServices k8s版本的服务发现和consul等不同，这里作为参数的serviceName不对应k8s的service的name，
//// 而是对应k8s的service的label名称。因为k8s中，可能有很多个service有同一个label，
//// 所以这里同一个serviceName，可能对应很多个k8s的service。而这个方法会把k8s中，所有带有serviceName名称的label的service找到，
//// 然后把其Endpoint中的所有pod的元信息（包括其提供服务的地址）封装成[]*ServiceInstance返回
//func (k *K8sDiscoveryClient) DiscoverServices(serviceName string) []*ServiceInstance {
//	labelName := serviceName
//	klog.Tracef("start to k8s DiscoverServices")
//	instanceList, ok := k.instancesMap.Load(labelName)
//	if ok {
//		return instanceList.([]*ServiceInstance)
//	}
//	k.mutex.Lock()
//	defer k.mutex.Unlock()
//	instanceList, ok = k.instancesMap.Load(labelName)
//	if ok {
//		return instanceList.([]*ServiceInstance)
//	}
//
//	ret := k.getAndStoreInstancesFromServicesByLabel(labelName)
//	go func() {
//		watcher, err := k.clientSet.CoreV1().Services(k.namespace).Watch(context.TODO(), metav1.ListOptions{
//			LabelSelector: labelName,
//			Watch:         true,
//		})
//		if err != nil {
//			klog.Errorf("DiscoverServices watch error: %v", err)
//			return
//		}
//		defer watcher.Stop()
//
//		for event := range watcher.ResultChan() {
//			klog.Infof("discovery fetched event:%+v", event)
//
//			svc, ok := event.Object.(*corev1.Service)
//			if !ok {
//				continue
//			}
//
//			switch event.Type {
//			case watch.Added, watch.Modified:
//				var endpoints *corev1.Endpoints
//				var err error
//
//				// 重试获取 Endpoint，测试使用。。
//				for i := 0; i < 3; i++ {
//					endpoints, err = k.clientSet.CoreV1().Endpoints(k.namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
//					if err == nil {
//						break
//					}
//					klog.Tracef("Failed to get endpoints for service %s: %v in namespace:%v, retrying...", svc.Name, err, k.namespace)
//					time.Sleep(1 * time.Second)
//				}
//
//				if err != nil {
//					klog.Errorf("Failed to get endpoints for service %s: %v in namespace:%v after retries", svc.Name, err, k.namespace)
//					continue
//				}
//
//				instances := k.getInstancesFromEndpoints(labelName, endpoints, svc.Labels)
//				k.k8sLabel2K8sService2ServiceInstance[labelName][svc.Name] = instances
//				k.refreshServiceInstance2SyncMap(labelName)
//				klog.Infof("discovery refreshServiceInstance2SyncMap:%v", k.k8sLabel2K8sService2ServiceInstance)
//			case watch.Deleted:
//				delete(k.k8sLabel2K8sService2ServiceInstance[labelName], svc.Name)
//				k.refreshServiceInstance2SyncMap(labelName)
//				klog.Infof("discovery refreshServiceInstance2SyncMap:%v", k.k8sLabel2K8sService2ServiceInstance)
//			default:
//				klog.Errorf("miss handle event type:%v", event.Type)
//			}
//		}
//	}()
//
//	return ret
//}
//
//func (k *K8sDiscoveryClient) getAndStoreInstancesFromServicesByLabel(labelName string) []*ServiceInstance {
//	services, err := k.clientSet.CoreV1().Services(k.namespace).List(context.TODO(), metav1.ListOptions{
//		LabelSelector: labelName,
//	})
//	if err != nil {
//		klog.Errorf("Failed to list services: %v", err)
//		return nil
//	}
//
//	var ret []*ServiceInstance
//	for _, svc := range services.Items {
//		endpoints, err := k.clientSet.CoreV1().Endpoints(k.namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
//		if err != nil {
//			klog.Errorf("Failed to get endpoints for service %s: %v", svc.Name, err)
//			continue
//		}
//
//		instanceFromEndpoint := k.getInstancesFromEndpoints(labelName, endpoints, svc.Labels)
//		ret = append(ret, instanceFromEndpoint...)
//		k.k8sLabel2K8sService2ServiceInstance[labelName] = make(map[string][]*ServiceInstance)
//		k.k8sLabel2K8sService2ServiceInstance[labelName][svc.Name] = instanceFromEndpoint
//	}
//
//	k.refreshServiceInstance2SyncMap(labelName)
//	return ret
//}
//
//// getInstancesFromEndpoints 从EndPoints信息中，解析出ServiceInstance信息
//func (k *K8sDiscoveryClient) getInstancesFromEndpoints(labelName string, endpoints *corev1.Endpoints,
//	labels map[string]string) []*ServiceInstance {
//	extraConfig, protoc, err := extractExtraConfigAndProtoc(labels)
//	if err != nil {
//		//todo 删掉
//		panic(err)
//	}
//
//	var instances []*ServiceInstance
//	for _, subset := range endpoints.Subsets {
//		for _, address := range subset.Addresses {
//			podName := address.TargetRef.Name
//			if podName == "" {
//				klog.Warnf("find podName empty:%+v", address)
//				continue
//			}
//
//			instance := &ServiceInstance{
//				ServiceName: labelName,
//				InstanceId:  podName, //k8s作为服务发现，使用podName作为其instanceID
//				ServiceServeConf: ServiceServeConf{
//					Protoc: protoc,
//					Host:   address.IP,
//					Port:   int(subset.Ports[0].Port), //这里直接用endpoint的第一个端口了，以后有需求再拓展
//				},
//				ExtraConfig: extraConfig,
//			}
//			instances = append(instances, instance)
//		}
//	}
//
//	return instances
//}
//
//// extractExtraConfigAndProtoc 从某个service对应的Endpoint中，解析出这个service的extraConfig和protoc这两个信息
//func extractExtraConfigAndProtoc(labels map[string]string) (string, ProtocType, error) {
//	var (
//		protoc ProtocType
//	)
//
//	//如果labels中没有protoc字段，报错
//	protocStr, ok := labels[k8sYamlLabelProtocFieldName]
//	if !ok {
//		return "", "", fmt.Errorf("can not find proto in labels:%v", labels)
//	} else {
//		switch ProtocType(protocStr) {
//		case ProtocTypeGrpc:
//			protoc = ProtocTypeGrpc
//		case ProtocTypeHttp:
//			protoc = ProtocTypeHttp
//		default:
//			return "", "", fmt.Errorf("can not decode proto type:%v", protocStr)
//		}
//	}
//
//	extraConfig := labels[k8sYamlLabelExtraConfigFieldName]
//	return extraConfig, protoc, nil
//}
//
//func (k *K8sDiscoveryClient) refreshServiceInstance2SyncMap(labelName string) {
//	tmp := make([]*ServiceInstance, 0, len(k.k8sLabel2K8sService2ServiceInstance[labelName]))
//
//	for _, instance := range k.k8sLabel2K8sService2ServiceInstance[labelName] {
//		tmp = append(tmp, instance...)
//	}
//
//	k.instancesMap.Store(labelName, tmp)
//}
//
//var _ DiscoverClient = (*K8sDiscoveryClient)(nil)
