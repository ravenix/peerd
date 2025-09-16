package kubernetes

import (
	"context"

	"github.com/ravenix/peerd/pkg/explorer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type resourceExplorer struct {
	explorer.Explorer

	k8sClient     *kubernetes.Clientset
	labelSelector string
	fieldSelector string

	listResources   func(context.Context) ([]any, error)
	exploreResource func(context.Context, any) *explorer.Discovery
}

type resourceExplorerConfig struct {
	ApiServer     string `yaml:"api_server"`
	CAFile        string `yaml:"ca_file"`
	TokenFile     string `yaml:"token_file"`
	LabelSelector string `yaml:"label_selector"`
	FieldSelector string `yaml:"field_selector"`
}

func newResourceExplorer(config *resourceExplorerConfig, listResources func(context.Context) ([]any, error), exploreResource func(context.Context, any) *explorer.Discovery) (*resourceExplorer, error) {
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

	return &resourceExplorer{
		k8sClient:     client,
		labelSelector: config.LabelSelector,
		fieldSelector: config.FieldSelector,

		listResources:   listResources,
		exploreResource: exploreResource,
	}, nil
}

func (e *resourceExplorer) Run(ctx context.Context) error {
	return nil
}

func (e *resourceExplorer) Explore(ctx context.Context, dh explorer.DiscoveryHandler) error {
	resources, err := e.listResources(ctx)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		if dis := e.exploreResource(ctx, resource); dis != nil {
			dh.Discovered(dis)
		}
	}

	return nil
}

func (e *resourceExplorer) newListOptions() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: e.labelSelector,
		FieldSelector: e.fieldSelector,
	}
}
