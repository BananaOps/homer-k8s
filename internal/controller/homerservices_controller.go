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
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/go-logr/logr"
	homerv1alpha1 "github.com/jplanckeel/homer-k8s/api/v1alpha1"
	homerconfig "github.com/jplanckeel/homer-k8s/pkg/config"
)

// Define logger
var logger logr.Logger
var logContext []interface{} = []interface{}{"controller", "homerservices", "controllerGroup", "homer.bananaops.io", "controllerKind", "HomerServices"}

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

func (r *HomerServicesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Get all CRD HomerServices
	allServices, error := getAllHomerServices(ctx, r)
	if error != nil {
		fmt.Println(error, "unable to fetch HomerServicesList")
		return ctrl.Result{}, error
	}
	pages, groups := splitServicesPerPage(allServices)

	localConfigs := map[string]homerconfig.HomerConfig{}
	files, _ := os.ReadDir("/assets")
	for _, f := range files {
		var localConfig homerconfig.HomerConfig
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") {
			file, _ := os.ReadFile("/assets/" + f.Name())
			err := yaml.Unmarshal(file, &localConfig)
			if err != nil {
				logger.Error(err, "error:")
			}
			localConfigs[f.Name()] = localConfig
		}
	}

	var globalConfig homerconfig.HomerConfig
	fileglobalConfig, _ := os.ReadFile("/config/global_config.yml")
	err := yaml.Unmarshal(fileglobalConfig, &globalConfig)
	if err != nil {
		logger.Error(err, "error:")
	}

	for page := range pages {
		config := globalConfig
		config.Services = sortServicesPerItemsLength(mergeGroupWithSameName(append(config.Services, groups[page]...)))

		for p, prettyP := range pages {
			if p != page {
				link := homerconfig.Link{
					Name: prettyP,
					Icon: "fas fa-list",
					URL:  "#" + p,
				}
				config.Links = append(config.Links, link)
			}
		}
		d, _ := yaml.Marshal(config)
		if !reflect.DeepEqual(config.Services, localConfigs[page].Services) {
			if page == "" {
				err = os.WriteFile("/assets/config.yml", d, 0600)
			} else {
				err = os.WriteFile("/assets/"+page+".yml", d, 0600)
			}
			if err != nil {
				logger.Error(err, "error:")
			}
			logger.Info("Homer Config "+page+" Updated", logContext...)
		}
	}

	for page := range localConfigs {
		if _, ok := pages[page]; !ok {
			err := os.Remove("/assets/" + page + ".yml")
			if err != nil {
				logger.Info("Homer Config "+page+" Removed", logContext...)
			} else {
				logger.Error(err, "Homer Config "+page+" Not Removed. Error:")
			}
		}

	}

	return ctrl.Result{}, nil
}

func splitServicesPerPage(allServices *homerv1alpha1.HomerServicesList) (map[string]string, map[string][]homerv1alpha1.Group) {
	pages := map[string]string{}
	groups := map[string][]homerv1alpha1.Group{}
	for _, service := range allServices.Items {
		if _, ok := groups[service.Spec.Page]; !ok {
			groups[service.Spec.Page] = []homerv1alpha1.Group{}
		}
		groups[service.Spec.Page] = append(groups[service.Spec.Page], service.Spec.Groups...)
		pages[strings.ReplaceAll(strings.ToLower(service.Spec.Page), " ", "-")] = service.Spec.Page
	}
	return pages, groups
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

// Init logger slog for json and output to stdout
func init() {
	opts := zap.Options{
		Development: false,
	}
	logger = zap.New(zap.UseFlagOptions(&opts))
}

func mergeGroupWithSameName(g []homerv1alpha1.Group) []homerv1alpha1.Group {
	groups := []homerv1alpha1.Group{}

	for _, g1 := range g {
		found := false
		for i, g2 := range groups {
			if g1.Name == g2.Name {
				groups[i].Items = append(groups[i].Items, g1.Items...)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, g1)
		}
	}

	return groups
}

func sortServicesPerItemsLength(g []homerv1alpha1.Group) []homerv1alpha1.Group {
	for i := range g {
		for j := range g {
			if len(g[i].Items) < len(g[j].Items) {
				g[i], g[j] = g[j], g[i]
			}
		}
	}

	return g
}
