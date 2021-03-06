/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"log"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "github.com/owenchenxy/kubehosts/api/v1"
	"github.com/owenchenxy/kubehosts/controllers/operation"
	"github.com/owenchenxy/kubehosts/controllers/spec"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// KubehostsReconciler reconciles a Kubehosts object
type KubehostsReconciler struct {
	Client   client.Client
	Logger   logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.gitee.com,resources=kubehosts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.gitee.com,resources=kubehosts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.gitee.com,resources=kubehosts/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="storage.k8s.io",resources=storageclasses,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kubehosts object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile

func (r *KubehostsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	log.Print("=======================================================")
	log.Print("=======================================================")

	kubeHosts, err := r.getDeployedKubehostsResource(ctx, req)
	if err != nil {
		return ctrl.Result{}, nil
	}

	if kubeHosts.Spec.HostsConfigMap == "" {
		return ctrl.Result{}, nil
	}

	configMapName := kubeHosts.Spec.HostsConfigMap
	foundConfigMap := &core.ConfigMap{}

	err = r.Client.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: req.Namespace}, foundConfigMap)

	if err != nil {
		// If a configMap name is provided, then it must exist
		// You will likely want to create an Event for the user to understand why their reconcile is failing.
		log.Println("configmap get error")
		return ctrl.Result{}, err
	}

	// ???namespace???????????????label???pod??????
	podList := &core.PodList{}
	// ????????????????????????????????????pod
	opts := []client.ListOption{
		//client.InNamespace(req.Namespace), // ?????????????????????req???namespace???
		//client.MatchingLabelsSelector{labels.SelectorFromSet(kubeHosts.Spec.Lables)},
	}
	for label, value := range kubeHosts.Spec.Lables {
		opts = append(opts, client.MatchingLabels{label: value})
	}
	err = r.Client.List(ctx, podList, opts...)
	if err != nil {
		log.Println(err)
	}

	hostsContent := strings.Replace(foundConfigMap.Data["hosts"], "\n", "\\n", -1)
	for _, pod := range podList.Items {
		go operation.WritePodHosts(pod, hostsContent)
	}
	return ctrl.Result{}, nil
}

func (r *KubehostsReconciler) getDeployedKubehostsResource(ctx context.Context, req ctrl.Request) (*batchv1.Kubehosts, error) {
	// We sleep 1 second to let sufficient time to Kubernetes to update its system
	// so that when we will call the Get method below, we will receive the latest Kubehosts resource
	//time.Sleep(1 * time.Second)
	kubehosts := &batchv1.Kubehosts{}
	err := r.Client.Get(ctx, req.NamespacedName, kubehosts)
	if err == nil {
		return kubehosts, nil
	}
	return &batchv1.Kubehosts{}, err
}

func (r *KubehostsReconciler) findObjectsForPod(pod client.Object) []reconcile.Request {
	podLabelMatchedKubehosts := &batchv1.KubehostsList{}
	allKubehosts := &batchv1.KubehostsList{}

	listOps := &client.ListOptions{}
	err := r.Client.List(context.TODO(), allKubehosts, listOps)
	if err != nil {
		return []reconcile.Request{}
	}
	for _, kubehosts := range allKubehosts.Items {
		controllerLables := kubehosts.Spec.Lables
		podLables := pod.GetLabels()
		for ckey, cvalue := range controllerLables {
			for pkey, pvalue := range podLables {
				if ckey == pkey && cvalue == pvalue {
					podLabelMatchedKubehosts.Items = append(podLabelMatchedKubehosts.Items, kubehosts)
				}
			}
		}
	}
	//log.Println(podLabelMatchedKubehosts)
	requests := make([]reconcile.Request, len(podLabelMatchedKubehosts.Items))
	for i, item := range podLabelMatchedKubehosts.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// ????????????????????????????????????????????????kubehosts controller????????????reconcile????????????????????????????????????????????????reconcile??????
func (r *KubehostsReconciler) findObjectsForConfigMap(configMap client.Object) []reconcile.Request {
	attachedConfigKubehosts := &batchv1.KubehostsList{}
	listOps := &client.ListOptions{
		// hostsConfigMap??????
		FieldSelector: fields.OneTermEqualSelector(".spec.hostsConfigMap", configMap.GetName()),
		//Namespace:     configMap.GetNamespace(),
	}
	err := r.Client.List(context.TODO(), attachedConfigKubehosts, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedConfigKubehosts.Items))
	for i, item := range attachedConfigKubehosts.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubehostsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// ???mgr??????kubehosts???configmap??????
	// ?????????????????????????????????kubehosts??????????????????kubehosts-1, kubehosts-2
	// kubehosts-1/2???.spec.hostsConfigMap???hosts-config-1/2
	// ?????????findObjectsForConfigMap??????fields.OneTermEqualSelector?????????.spec.hostsConfigMap????????????hosts-config-1/2
	// ????????????hostsConfigMap???????????????????????????kubehosts

	if err := spec.IndexHostsConfigMap(mgr); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Kubehosts{}).
		// ????????????configmap??????????????????????????????configmap???????????????????????????configmap???????????????????????????findObjectsForConfigMap??????
		Watches(
			&source.Kind{Type: &core.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&source.Kind{Type: &core.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
