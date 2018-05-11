package framework

import (
	api "github.com/appscode/kubed/apis/kubed/v1alpha1"
)

func SnapshotterClusterConfig(backend *api.Backend) api.ClusterConfig {
	return api.ClusterConfig{
		Snapshotter: &api.SnapshotSpec{
			Backend:  *backend,
			Sanitize: true,
			Schedule: "@every 1m",
		},
	}
}

func ConfigMapSyncClusterConfig() api.ClusterConfig {
	return api.ClusterConfig{
		EnableConfigSyncer: true,
	}
}

func (f *Invocation) EventForwarderClusterConfig() api.ClusterConfig {
	return api.ClusterConfig{
		EventForwarder: &api.EventForwarderSpec{
			Rules: []api.PolicyRule{
				{
					Operations: []api.Operation{api.Create},
					Namespaces: []string{f.namespace},
					Resources: []api.GroupResources{
						{
							Group: "",
							Resources: []string{
								"events",
							},
						},
					},
				},
				{
					Operations: []api.Operation{api.Create},
					Namespaces: []string{f.namespace},
					Resources: []api.GroupResources{
						{
							Group: "",
							Resources: []string{
								"persistentvolumeclaims",
							},
						},
					},
				},
			},
		},
	}
}

func APIServerClusterConfig() api.ClusterConfig {
	return api.ClusterConfig{
		ClusterName: "minikube",
	}
}

func WebhookReceiver() []api.Receiver {
	return []api.Receiver{
		{
			To:       []string{"ops-alerts"},
			Notifier: "Webhook",
		},
	}
}

func ResetTestConfigFile() error {
	defaultClusterConfig := api.ClusterConfig{
		ClusterName: "minikube",
	}
	return defaultClusterConfig.Save(KubedTestConfigFileDir)
}
