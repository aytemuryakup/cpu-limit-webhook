package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"io/ioutil"
	//"reflect"
	"log"
	"net/http"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)
var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type webhook struct{}

func (vh *webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get webhook body with the admission review.
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		http.Error(w, "no body found", http.StatusBadRequest)
		return
	}

	ar := &admissionv1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, ar); err != nil {
		http.Error(w, "could not decode the admission review from the request", http.StatusBadRequest)
		return
	}
	fmt.Println("New request ", ar.Request.UID, ar.Request.Kind, ar.Request.Operation, ar.Request.Namespace)
	if ar.Request.Kind.Kind != "Deployment" {
		fmt.Println("Resource type is not deployment, skipped", ar.Request.Kind)
		return
	}

	status := "Success"
	deploy := &appsv1.Deployment{}
	if _, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, deploy); err != nil {
		http.Error(w, "could not decode admission request object", http.StatusBadRequest)
		return
	}
	for _, container := range deploy.Spec.Template.Spec.Containers {
		limits := container.Resources.Limits
		requests := container.Resources.Requests

		cpuLimit := limits["cpu"]
		cpuRequest := requests["cpu"]
		fmt.Println("Container resources cpu ", cpuLimit, cpuRequest)

			limit, ok := cpuLimit.AsInt64()
			if !ok {
				limit_m := cpuLimit.String()
				limit_lenght := len(limit_m)
				if limit_lenght > 0 && limit_m[limit_lenght-1] == 'm' {
					limit_m = limit_m[:limit_lenght-1]
					// fmt.Println("limit_m degeri ", limit_m)
					limit_convert, erc := strconv.ParseInt(limit_m, 10, 64)
					if erc == nil {
						if limit_convert > 2000 {
							fmt.Println("CPU limit exceeded.")
							status = "Failure"
						}
					}
				}
				continue
			}
			if limit > 2 {
				fmt.Println("CPU limit exceeded.")
				status = "Failure"
			}
			request, ok := cpuRequest.AsInt64()
			if !ok {
				request_m := cpuRequest.String()
				request_lenght := len(request_m)
				if request_lenght > 0 && request_m[request_lenght-1] == 'm' {
					request_m = request_m[:request_lenght-1]
					request_convert, erc := strconv.ParseInt(request_m, 10, 64)
					if erc == nil {
						if request_convert > 2000 {
							fmt.Println("CPU request exceeded.")
							status = "Failure"
						}
					}
				}
				continue
			}
			if request > 2 {
				fmt.Println("CPU limit exceeded.")
				status = "Failure"
			}
	}
	allowed := true
	if status != "Success" {
		allowed = false
	}

	admissionResp := &admissionv1beta1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Status:  status,
			Message: "ok",
		},
	}

	// Forge the review response.
	aResponse := admissionv1beta1.AdmissionReview{
		Response: admissionResp,
	}
	resp, err := json.Marshal(aResponse)
	if err != nil {
		http.Error(w, "error marshaling to json admission review response", http.StatusInternalServerError)
		return
	}
	// Forge the HTTP response.
	// If the received admission review has failed mark the response as failed.
	if admissionResp.Result != nil && admissionResp.Result.Status == metav1.StatusFailure {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(resp); err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	srv := &http.Server{
		Addr:    ":8443",
		Handler: &webhook{},
		// TLSConfig: cfg,
	}
	fmt.Println("Start listen:8443")
	log.Fatal(srv.ListenAndServeTLS("certs/tls.crt", "certs/tls.key"))
}
