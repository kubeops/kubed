package framework

import (
	"sync"

	"github.com/appscode/kubed/pkg/operator"
	"k8s.io/client-go/rest"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	. "github.com/onsi/gomega"
	kcs "github.com/k8sdb/apimachinery/client/clientset"
	scs "github.com/appscode/stash/client/clientset"
	vcs "github.com/appscode/voyager/client/clientset"
	// pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	srch_cs "github.com/appscode/searchlight/client/clientset"
	"github.com/appscode/go/crypto/rand"
)

const (
	MaxRetry = 200
	NoRetry  = 1
)

type Framework struct {
	KubeConfig    *rest.Config
	KubeClient    clientset.Interface
	KubedOperator *operator.Operator
	Config        E2EConfig
	namespace     string
	Mutex         sync.Mutex
}

type Invocation struct {
	*Framework
	app string
}

func New() *Framework {
	testConfigs.validate()

	c, err := clientcmd.BuildConfigFromFlags(testConfigs.Master, testConfigs.KubeConfig)
	Expect(err).NotTo(HaveOccurred())

	return &Framework{
		KubeConfig: c,
		KubeClient: clientset.NewForConfigOrDie(c),
		namespace:  testConfigs.TestNamespace,
		Config:     testConfigs,
		KubedOperator: &operator.Operator{
			KubeClient:        clientset.NewForConfigOrDie(c),
			StashClient:       scs.NewForConfigOrDie(c),
			VoyagerClient:     vcs.NewForConfigOrDie(c),
			SearchlightClient: srch_cs.NewForConfigOrDie(c),
			KubeDBClient:      kcs.NewForConfigOrDie(c),
		},
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("kubed-e2e"),
	}
}

func (f *Invocation) App() string {
	return f.app
}