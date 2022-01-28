package spec

import (
	"context"

	batchv1 "github.com/owenchenxy/kubehosts/api/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
func IndexLables(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&batchv1.Kubehosts{},
		".spec.lables",
		func(rawObj client.Object) []string {
			// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
			kubeHosts := rawObj.(*batchv1.Kubehosts)
			if kubeHosts.Spec.Lables == nil {
				return nil
			}
			lablesString, err := json.Marshal(kubeHosts.Spec.Lables)
			if err != nil {
				return nil
			}
			return []string{string(lablesString)}
		}); err != nil {
		return err
	}
	return nil
}
*/
func IndexHostsConfigMap(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&batchv1.Kubehosts{},
		".spec.hostsConfigMap",
		func(rawObj client.Object) []string {
			// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
			kubeHosts := rawObj.(*batchv1.Kubehosts)
			if kubeHosts.Spec.HostsConfigMap == "" {
				return nil
			}
			return []string{kubeHosts.Spec.HostsConfigMap}
		}); err != nil {
		return err
	}
	return nil
}
