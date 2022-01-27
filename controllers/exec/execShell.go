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

package exec

import (
	"bytes"
	"io"
	"log"

	"github.com/owenchenxy/kubehosts/controllers/singleton"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecCmdInContainer(namespace string, podName string, containerName string, command string) (string, string, error) {
	clientSet := singleton.GetClient()
	config := singleton.GetConfig()
	stdoutString, stderrString, err := ExecCmd(clientSet, config, podName, containerName, namespace, command, nil)
	if err != nil {
		log.Println(err)
		return "", "", err
	}
	return stdoutString, stderrString, err
}

func ExecCmd(client kubernetes.Interface, config *restclient.Config, podName string, containerName string, namespace string,
	command string, stdin io.Reader) (string, string, error) {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}

	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", nil
	}
	return stdout.String(), stderr.String(), nil
}
