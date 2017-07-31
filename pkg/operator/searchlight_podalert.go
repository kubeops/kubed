package operator

import (
	"errors"
	"reflect"

	acrt "github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/util"
	"github.com/appscode/log"
	tapi "github.com/appscode/searchlight/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

// Blocks caller. Intended to be called as a Go routine.
func (op *Operator) WatchPodAlerts() {
	if !util.IsPreferredAPIResource(op.KubeClient, tapi.V1alpha1SchemeGroupVersion.String(), tapi.ResourceKindPodAlert) {
		log.Warningf("Skipping watching non-preferred GroupVersion:%s Kind:%s", tapi.V1alpha1SchemeGroupVersion.String(), tapi.ResourceKindPodAlert)
		return
	}

	defer acrt.HandleCrash()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return op.SearchlightClient.PodAlerts(apiv1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return op.SearchlightClient.PodAlerts(apiv1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&tapi.PodAlert{},
		op.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if res, ok := obj.(*tapi.PodAlert); ok {
					log.Infof("PodAlert %s@%s added", res.Name, res.Namespace)
					util.AssignTypeKind(res)

					if op.Config.APIServer.EnableSearchIndex {
						if err := op.SearchIndex.HandleAdd(obj); err != nil {
							log.Errorln(err)
						}
					}

					if op.Config.APIServer.EnableReverseIndex {
						if err := op.ReverseIndex.PodAlert.Add(res); err != nil {
							log.Errorln(err)
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if res, ok := obj.(*tapi.PodAlert); ok {
					log.Infof("PodAlert %s@%s deleted", res.Name, res.Namespace)
					util.AssignTypeKind(res)

					if op.Config.APIServer.EnableSearchIndex {
						if err := op.SearchIndex.HandleDelete(obj); err != nil {
							log.Errorln(err)
						}
					}
					if op.TrashCan != nil {
						op.TrashCan.Delete(res.TypeMeta, res.ObjectMeta, obj)
					}

					if op.Config.APIServer.EnableReverseIndex {
						if err := op.ReverseIndex.PodAlert.Delete(res); err != nil {
							log.Errorln(err)
						}
					}
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldRes, ok := old.(*tapi.PodAlert)
				if !ok {
					log.Errorln(errors.New("Invalid PodAlert object"))
					return
				}
				newRes, ok := new.(*tapi.PodAlert)
				if !ok {
					log.Errorln(errors.New("Invalid PodAlert object"))
					return
				}
				util.AssignTypeKind(oldRes)
				util.AssignTypeKind(newRes)

				if op.Config.APIServer.EnableSearchIndex {
					op.SearchIndex.HandleUpdate(old, new)
				}

				if op.Config.APIServer.EnableReverseIndex {
					if err := op.ReverseIndex.PodAlert.Update(oldRes, newRes); err != nil {
						log.Errorln(err)
					}
				}

				if op.TrashCan != nil && op.Config.RecycleBin.HandleUpdates {
					if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
						!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
						!reflect.DeepEqual(oldRes.Spec, newRes.Spec) {
						op.TrashCan.Update(newRes.TypeMeta, newRes.ObjectMeta, old, new)
					}
				}
			},
		},
	)
	ctrl.Run(wait.NeverStop)
}
