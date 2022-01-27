package operation

import (
	"fmt"

	"github.com/owenchenxy/kubehosts/controllers/exec"
	core "k8s.io/api/core/v1"
)

func WritePodHosts(pod core.Pod, hostsContent string) {
	for _, container := range pod.Spec.Containers {
		go writeContainerHosts(pod, container, hostsContent)
	}
}

func writeContainerHosts(pod core.Pod, container core.Container, hostsContent string) {
	exec.ExecCmdInContainer(
		pod.Namespace,
		pod.Name,
		container.Name,
		fmt.Sprintf(`echo -e "%s" > /etc/hosts`, hostsContent),
	)
}
