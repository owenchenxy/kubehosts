[Kubehosts] is a Kubernetes operator allowing to write self defined /etc/hosts file to a set of pods matching the specified labels.

**Features**

* It can sense the change of the config map contents. If the config map specified in the kubehosts' '.spec.hostsConfigMap' field changes, the controller will write write the /etc/hosts file for related pods.

* It can sense the change of pods. If some pods are created/updated/deleted,  the controller will write the /etc/hosts file for related pods.


**How does Kubehosts determine which pod to process?**

It checks the labels of the pod, if one or more label matches one of the kubehosts' '.spec.lables', the operator will write the /etc/hosts file for the pod.

**Getting started**

TODO

