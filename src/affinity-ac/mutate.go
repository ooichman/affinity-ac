package main

import(
	"net/http"
	"os"
	"io/ioutil"
	"fmt"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func isKubeNamespace(ns string) bool {
	return ns == metav1.NamespacePublic || ns == metav1.NamespaceSystem
}

type patchOperation struct {
	Op    string    `json:"op"`
	Path  string    `json:"path"`
	Value string    `json:"Value"`
}

func (mac *myServerHandler) mutserve(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(os.Stdout, "Starting Mutation request\n")

	var Body []byte

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			Body = data
		}
	} 

	if len(Body) == 0 {
		fmt.Fprintf(os.Stdout, "Unable to retrieve the Body from the API\n")
		http.Error(w, "Unable to retrieve the Body from the API", http.StatusBadRequest)
	}

	fmt.Fprintf(os.Stdout, "Request Received\n")

	if r.URL.Path != "/mutate" {
		fmt.Fprintf(os.Stderr, "Not a valid URL Path\n")
		http.Error(w, "Not a valid URL Path", http.StatusBadRequest)
	}

	arRequest := &admissionv1.AdmissionReview{}
	if err := json.Unmarshal(Body, arRequest); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to Unmarshal the Body Request: %v\n", err)
		http.Error(w, "Unable to Unmarshal the Body Request", http.StatusBadRequest)
		return
	}

	raw := arRequest.Request.Object.Raw
	obj := corev1.Pod{}

	if !isKubeNamespace(arRequest.Request.Namespace) {

		if err := json.Unmarshal(raw, &obj); err != nil {
			fmt.Fprintf(os.Stderr, "Error , unable to Deserializing Pod\n")
			http.Error(w, "Error, unable to Deserializing Pod", http.StatusBadRequest)
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error , unauthorized Namespace\n")
		http.Error(w, "Error , unauthorized Namespace", http.StatusBadRequest )
		return
	}

	arResponse := admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{
			UID: arRequest.Request.UID,
		},
	}

	podMeta := metav1.ObjectMeta{}
	metaraw := arRequest.Request.Options.Raw

	if err := json.Unmarshal(metaraw,&podMeta); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to Unmarshal the Pod Metadata\n")
		http.Error(w, "Unable to Unmarshal the Pod Metadata\n", http.StatusBadRequest)
	}

	labels := podMeta.Labels

	var patches []patchOperation

	if obj.Spec.Affinity == nil {
		fmt.Fprintf(os.Stdout, "No Affinity/AntiAffinity is define, Starting Mutation\n")

		patches = append(patches, patchOperation{
			Op: "Add",
			Path: "/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/0/labelSelector/matchExpressions/0/key",
			Value: "app",
		})

		patches = append(patches, patchOperation{
			Op: "Add",
			Path: "/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/0/labelSelector/matchExpressions/0/operator",
			Value: "In",
		})

		if labels["app"] != "" {
			patches = append(patches, patchOperation{
				Op: "Add",
				Path: "/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/0/labelSelector/matchExpressions/0/values",
				Value: labels["app"],
			})
		}

		patches = append(patches, patchOperation{
			Op: "Add",
			Path: "/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/0/topologyKey",
			Value: "kubernetes.io/hostname",
		})
	}

	fmt.Fprintf(os.Stdout, "the Json Is : \"%s\"\n", patches)
	patchBytes, err := json.Marshal(patches)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode the Patches")
		http.Error(w, "Unable to encode the Patches", http.StatusBadRequest)
		return
	}

	v1JSONPatch := admissionv1.PatchTypeJSONPatch
	arResponse.APIVersion = "admission.k8s.io/v1"
	arResponse.Kind = arRequest.Kind
	arResponse.Response.Allowed = true
	arResponse.Response.Patch = patchBytes
	arResponse.Response.PatchType = &v1JSONPatch

	resp , rerr := json.Marshal(arResponse)

	if rerr != nil {
		fmt.Fprintf(os.Stderr, "Unable to Marshal the Response: %v\n", rerr)
		http.Error(w, "Unable to Marshal the Response", http.StatusBadRequest)
	}

	if _, werr := w.Write(resp); werr != nil {
		fmt.Fprintf(os.Stderr , "Can't write response: %v", werr)
		http.Error(w, "Can't write a response", http.StatusBadRequest)
	}
}