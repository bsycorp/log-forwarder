package main

import "time"

type PodList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
	} `json:"metadata"`
	Items []Pod `json:"items"`
}

type Pod struct {
	Metadata struct {
		Name              string    `json:"name"`
		GenerateName      string    `json:"generateName"`
		Namespace         string    `json:"namespace"`
		SelfLink          string    `json:"selfLink"`
		UID               string    `json:"uid"`
		ResourceVersion   string    `json:"resourceVersion"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Labels            struct {
			ControllerRevisionHash string `json:"controller-revision-hash"`
			K8SApp                 string `json:"k8s-app"`
			PodTemplateGeneration  string `json:"pod-template-generation"`
		} `json:"labels"`
		Annotations struct {
			SumologicTrustedTimestamp string `json:"sumologic.com/trustedTimestamp"`
		} `json:"annotations"`
		OwnerReferences []struct {
			APIVersion         string `json:"apiVersion"`
			Kind               string `json:"kind"`
			Name               string `json:"name"`
			UID                string `json:"uid"`
			Controller         bool   `json:"controller"`
			BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
		} `json:"ownerReferences"`
	} `json:"metadata"`
	Spec struct {
		Volumes []struct {
			Name     string `json:"name"`
			HostPath struct {
				Path string `json:"path"`
				Type string `json:"type"`
			} `json:"hostPath,omitempty"`
			ConfigMap struct {
				Name        string `json:"name"`
				DefaultMode int    `json:"defaultMode"`
			} `json:"configMap,omitempty"`
			Secret struct {
				SecretName  string `json:"secretName"`
				DefaultMode int    `json:"defaultMode"`
			} `json:"secret,omitempty"`
		} `json:"volumes"`
		Containers []struct {
			Name  string `json:"name"`
			Image string `json:"image"`
			Env   []struct {
				Name      string `json:"name"`
				Value     string `json:"value,omitempty"`
				ValueFrom struct {
					FieldRef struct {
						APIVersion string `json:"apiVersion"`
						FieldPath  string `json:"fieldPath"`
					} `json:"fieldRef"`
				} `json:"valueFrom,omitempty"`
			} `json:"env"`
			Resources struct {
				Requests struct {
					CPU string `json:"cpu"`
				} `json:"requests"`
			} `json:"resources"`
			VolumeMounts []struct {
				Name      string `json:"name"`
				ReadOnly  bool   `json:"readOnly,omitempty"`
				MountPath string `json:"mountPath"`
			} `json:"volumeMounts"`
			LivenessProbe struct {
				HTTPGet struct {
					Path   string `json:"path"`
					Port   int    `json:"port"`
					Scheme string `json:"scheme"`
				} `json:"httpGet"`
				InitialDelaySeconds int `json:"initialDelaySeconds"`
				TimeoutSeconds      int `json:"timeoutSeconds"`
				PeriodSeconds       int `json:"periodSeconds"`
				SuccessThreshold    int `json:"successThreshold"`
				FailureThreshold    int `json:"failureThreshold"`
			} `json:"livenessProbe,omitempty"`
			ReadinessProbe struct {
				HTTPGet struct {
					Path   string `json:"path"`
					Port   int    `json:"port"`
					Scheme string `json:"scheme"`
				} `json:"httpGet"`
				TimeoutSeconds   int `json:"timeoutSeconds"`
				PeriodSeconds    int `json:"periodSeconds"`
				SuccessThreshold int `json:"successThreshold"`
				FailureThreshold int `json:"failureThreshold"`
			} `json:"readinessProbe,omitempty"`
			TerminationMessagePath   string `json:"terminationMessagePath"`
			TerminationMessagePolicy string `json:"terminationMessagePolicy"`
			ImagePullPolicy          string `json:"imagePullPolicy"`
			SecurityContext          struct {
				Privileged bool `json:"privileged"`
			} `json:"securityContext,omitempty"`
			Command []string `json:"command,omitempty"`
		} `json:"containers"`
		RestartPolicy                 string `json:"restartPolicy"`
		TerminationGracePeriodSeconds int    `json:"terminationGracePeriodSeconds"`
		DNSPolicy                     string `json:"dnsPolicy"`
		ServiceAccountName            string `json:"serviceAccountName"`
		ServiceAccount                string `json:"serviceAccount"`
		NodeName                      string `json:"nodeName"`
		HostNetwork                   bool   `json:"hostNetwork"`
		SecurityContext               struct {
		} `json:"securityContext"`
		SchedulerName string `json:"schedulerName"`
		Tolerations   []struct {
			Key      string `json:"key,omitempty"`
			Operator string `json:"operator"`
			Effect   string `json:"effect,omitempty"`
		} `json:"tolerations"`
	} `json:"spec"`
	Status struct {
		Phase      string `json:"phase"`
		Conditions []struct {
			Type               string      `json:"type"`
			Status             string      `json:"status"`
			LastProbeTime      interface{} `json:"lastProbeTime"`
			LastTransitionTime time.Time   `json:"lastTransitionTime"`
		} `json:"conditions"`
		HostIP            string    `json:"hostIP"`
		PodIP             string    `json:"podIP"`
		StartTime         time.Time `json:"startTime"`
		ContainerStatuses []struct {
			Name  string `json:"name"`
			State struct {
				Running struct {
					StartedAt time.Time `json:"startedAt"`
				} `json:"running"`
			} `json:"state"`
			LastState struct {
			} `json:"lastState"`
			Ready        bool   `json:"ready"`
			RestartCount int    `json:"restartCount"`
			Image        string `json:"image"`
			ImageID      string `json:"imageID"`
			ContainerID  string `json:"containerID"`
		} `json:"containerStatuses"`
		QosClass string `json:"qosClass"`
	} `json:"status"`
}
