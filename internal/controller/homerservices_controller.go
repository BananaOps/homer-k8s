/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	homerv1alpha1 "github.com/jplanckeel/homer-k8s/api/v1alpha1"
)

// HomerServicesReconciler reconciles a HomerServices object
type HomerServicesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HomerServices object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile

type Config struct {
	Services []homerv1alpha1.Group `yaml:"services"`
}

func (r *HomerServicesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var localConfig Config
	var config Config
	allServices, error := getAllHomerServices(ctx, r)
	if error != nil {
		fmt.Println(error, "unable to fetch HomerServicesList")
		return ctrl.Result{}, error
	}
	for _, service := range allServices.Items {
		config.Services = append(config.Services, service.Spec.Groups...)
	}


	file, _ := os.ReadFile("/assets/config.yml")
	err := yaml.Unmarshal(file, &localConfig)
	if err != nil {
		fmt.Println("error: %v", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("error: %v", err)
	}

	
	if !reflect.DeepEqual(config, localConfig) {
		err = os.WriteFile("/assets/config.yml", data, 0644)
		if err != nil {
			fmt.Println("error: %v", err)
		}

		fmt.Println("config.yaml updated")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HomerServicesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homerv1alpha1.HomerServices{}).
		Complete(r)
}

// Get all HomerServices
func getAllHomerServices(ctx context.Context, r *HomerServicesReconciler) (*homerv1alpha1.HomerServicesList, error) {
	var listService homerv1alpha1.HomerServicesList
	if err := r.List(ctx, &listService); err != nil {
		return nil, err
	}

	return &listService, nil
}
