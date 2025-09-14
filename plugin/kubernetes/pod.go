package kubernetes

import (
	"context"
	"net"

	"github.com/ravenix/peerd/pkg/explorer"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type podExplorer struct {
	k8sClient     *kubernetes.Clientset
	namespace     string
	labelSelector string
	port          uint16
}

type podExplorerConfig struct {
	ApiServer     string `yaml:"api_server"`
	CAFile        string `yaml:"ca_file"`
	TokenFile     string `yaml:"token_file"`
	Namespace     string `yaml:"namespace"`
	LabelSelector string `yaml:"label_selector"`
	Port          uint16 `yaml:"port"`
}

func podExplorerInitializer(yamlConfig *yaml.Node) (explorer.Explorer, error) {
	var config podExplorerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newpodExplorer(&config)
}

func newpodExplorer(config *podExplorerConfig) (*podExplorer, error) {
	k8sConfig := &rest.Config{
		Host:            config.ApiServer,
		BearerTokenFile: config.TokenFile,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: config.CAFile,
		},
	}

	client, err := kubernetes.NewForConfig(k8sConfig)

	if err != nil {
		return nil, err
	}

	return &podExplorer{
		k8sClient:     client,
		namespace:     config.Namespace,
		labelSelector: config.LabelSelector,
		port:          config.Port,
	}, nil
}

func (e *podExplorer) Run(ctx context.Context) error {
	return nil
}

func (e *podExplorer) Explore(ctx context.Context, dh explorer.DiscoveryHandler) error {
	pods, err := e.k8sClient.CoreV1().Pods(e.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: e.labelSelector,
	})

	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if pod.Status.PodIP != "" {
			dis := &explorer.Discovery{
				Port: uint16(e.port),
			}
			podIP := net.ParseIP(pod.Status.PodIP)

			if podIP.To4() != nil {
				dis.IPv4Addr = podIP
			} else {
				dis.IPv6Addr = podIP
			}

			dh.Discovered(dis)
		}
	}

	return nil
}
