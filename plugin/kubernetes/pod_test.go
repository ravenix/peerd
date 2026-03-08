package kubernetes

import "testing"

func TestNewPodExplorerSetsNamespace(t *testing.T) {
	e, err := newpodExplorer(&podExplorerConfig{
		resourceExplorerConfig: resourceExplorerConfig{
			ApiServer: "https://127.0.0.1",
		},
		Namespace: "router-system",
		PodPort:   179,
	})
	if err != nil {
		t.Fatalf("unexpected error creating pod explorer: %v", err)
	}

	if e.namespace != "router-system" {
		t.Fatalf("unexpected namespace: %q", e.namespace)
	}
}

