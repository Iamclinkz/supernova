package discovery

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudwego/kitex/pkg/klog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sDiscoveryClient struct {
	clientset      *kubernetes.Clientset
	httpServer     *http.Server
	namespace      string
	labelSelector  string
	instancePrefix string
}

func NewK8sDiscoveryClient(namespace, labelSelector, instancePrefix string) (ExecutorDiscoveryClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", "/path/to/kubeconfig")
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sDiscoveryClient{
		clientset:      clientset,
		namespace:      namespace,
		labelSelector:  labelSelector,
		instancePrefix: instancePrefix,
	}, nil
}

func (k *K8sDiscoveryClient) Register(instance *ExecutorServiceInstance) error {
	healthCheckPort := 8080 // You can customize this port

	k.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", healthCheckPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
	}

	go func() {
		if err := k.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("Error starting HTTP server for health check: %v", err)
		}
	}()

	// You may need to customize the readinessProbe and livenessProbe settings according to your needs
	readinessProbe := &corev1.Probe{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/health",
			Port: intstr.FromInt(healthCheckPort),
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
	}

	livenessProbe := &corev1.Probe{
		HTTPGet: &corev1.HTTPGetAction{
			Path: "/health",
			Port: intstr.FromInt(healthCheckPort),
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
	}

	// Update your Pod's readinessProbe and livenessProbe
	_, err := k.clientset.CoreV1().Pods(k.namespace).Patch(context.TODO(), instance.InstanceId, types.StrategicMergePatchType, []byte(fmt.Sprintf(`{"spec": {"containers": [{"name": "your-container-name", "readinessProbe": %v, "livenessProbe": %v}]}}`, readinessProbe, livenessProbe)), metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("Failed to update Pod's probes: %v", err)
		return err
	}

	return nil
}

func (k *K8sDiscoveryClient) DeRegister(instanceId string) error {
	if k.httpServer != nil {
		if err := k.httpServer.Shutdown(context.Background()); err != nil {
			klog.Errorf("Error stopping HTTP server for health check: %v", err)
		}
	}

	return nil
}

func (k *K8sDiscoveryClient) DiscoverServices() []*ExecutorServiceInstance {
	services, err := k.clientset.CoreV1().Services(k.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: k.labelSelector,
	})
	if err != nil {
		klog.Errorf("Failed to list services: %v", err)
		return nil
	}

	var instances []*ExecutorServiceInstance
	for _, svc := range services.Items {
		endpoints, err := k.clientset.CoreV1().Endpoints(k.namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get endpoints for service %s: %v", svc.Name, err)
			continue
		}

		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				podName := address.TargetRef.Name
				if podName == "" {
					continue
				}

				instance := &ExecutorServiceInstance{
					InstanceId: podName,
					ExecutorServiceServeConf: ExecutorServiceServeConf{
						Protoc: yourProtocValue, // Replace with the correct value for Protoc
						Host:   address.IP,
						Port:   int(subset.Ports[0].Port),
					},
					Tags: extractTags(svc.Labels),
				}
				instances = append(instances, instance)
			}
		}
	}

	return instances
}

func extractTags(labels map[string]string) []string {
	var tags []string
	for key, value := range labels {
		if strings.HasPrefix(key, "X-Tag-") {
			tags = append(tags, value)
		}
	}
	return tags
}

var _ ExecutorDiscoveryClient = (*K8sDiscoveryClient)(nil)
