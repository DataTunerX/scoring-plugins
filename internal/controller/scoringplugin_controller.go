/*
Copyright 2023.

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
	"encoding/json"
	"os"
	"path/filepath"

	extensionv1beta1 "github.com/DataTunerX/meta-server/api/extension/v1beta1"
	"github.com/DataTunerX/scoring-plugins/pkg/config"
	"github.com/DataTunerX/utility-server/logging"
	"github.com/DataTunerX/utility-server/parser"
	"github.com/qiniu/x/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScoringPluginReconciler reconciles a ScoringPlugin object
type ScoringPluginReconciler struct {
	client.Client
	Log    logging.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=extension.datatunerx.io,resources=scoringplugins,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extension.datatunerx.io,resources=scoringplugins/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extension.datatunerx.io,resources=scoringplugins/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScoringPlugin object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile

// Reconcile reads that state of the cluster for a Scoring object and makes changes based on the state read
// and what is in the Scoring.Spec
func (r *ScoringPluginReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Reconciling Scoring")

	// Fetch the Scoring instance
	var scoring extensionv1beta1.Scoring
	if err := r.Get(ctx, req.NamespacedName, &scoring); err != nil {
		r.Log.Errorf("unable to fetch Scoring: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the ScoringPlugin instance used by the Scoring
	var scoringPlugin extensionv1beta1.ScoringPlugin
	if scoring.Spec.Plugin.LoadPlugin {
		scoringPluginName := scoring.Spec.Plugin.Name
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: config.GetDatatunerxSystemNamespace(),
			Name:      scoringPluginName,
		}, &scoringPlugin); err != nil {
			r.Log.Errorf("unable to fetch ScoringPlugin: %v", err)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		// Build the path to the plugin YAML file
		pluginPath := filepath.Join("plugins", scoringPlugin.Spec.Provider, scoringPlugin.Spec.ScoringClass, "plugin.yaml")
		// Apply the plugin YAML file
		if err := r.applyYAML(ctx, pluginPath, &scoring); err != nil {
			r.Log.Errorf("unable to apply plugin YAML %v: %v", pluginPath, err)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// applyYAML reads a YAML file, replaces placeholders with environment variable values, and applies its content to the Kubernetes cluster
func (r *ScoringPluginReconciler) applyYAML(ctx context.Context, path string, scoring *extensionv1beta1.Scoring) error {

	r.Log.Infof("Applying plugin YAML %v", path)
	// Read the YAML file content
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		r.Log.Errorf("unable to read plugin YAML file: %v", err)
		return err
	}

	// Convert the file content to a string
	yamlStr := string(yamlFile)

	// Replace placeholders with environment variable values and run-time parameters defined in the dataset
	replacedYamlStr, err := r.replacePlaceholders(yamlStr, scoring)
	if err != nil {
		r.Log.Errorf("unable to replace placeholders in YAML: %v", err)
		return err
	}

	// Convert the updated YAML string back to a byte slice
	yamlFile = []byte(replacedYamlStr)

	// Decode the YAML into an unstructured.Unstructured object
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	unstructuredObj := &unstructured.Unstructured{}
	_, _, err = decUnstructured.Decode(yamlFile, nil, unstructuredObj)
	if err != nil {
		r.Log.Errorf("unable to decode YAML into Unstructured: %v", err)
		return err
	}

	// Set the namespace and owner reference
	unstructuredObj.SetNamespace(scoring.GetNamespace())
	if err := ctrl.SetControllerReference(scoring, unstructuredObj, r.Scheme); err != nil {
		r.Log.Errorf("unable to set controller reference: %v", err)
		return err
	}

	// Apply the unstructured object using the client
	if err := r.applyClient(ctx, unstructuredObj); err != nil {
		r.Log.Errorf("unable to apply Unstructured object: %v", err)
		return err
	}

	return nil
}

// replacePlaceholders replaces a specific placeholder in the YAML file with the value from an environment variable
func (r *ScoringPluginReconciler) replacePlaceholders(yamlStr string, scoring *extensionv1beta1.Scoring) (string, error) {

	// Parameters holding the unmarshaled parameters
	var parameters map[string]interface{}

	// Unmarshal the parameters
	err := json.Unmarshal([]byte(scoring.Spec.Plugin.Parameters), &parameters)
	if err != nil {
		r.Log.Errorf("unable to unmarshal plugin parameters: %v", err)
		return "", err
	}

	// Add the required fields defined in the plugin standard to parameters
	parameters["completeNotifyUrl"] = config.GetCompleteNotifyURL()
	parameters["inferenceService"] = scoring.Spec.InferenceService

	// Replace the value in template yaml
	replacedYamlStr, err := parser.ReplaceTemplate(yamlStr, parameters)
	if err != nil {
		r.Log.Errorf("unable to replace placeholders in YAML: %v", err)
		return "", err
	}

	return replacedYamlStr, nil
}

// applyClient applies or updates the given unstructured object in the cluster using the client
func (r *ScoringPluginReconciler) applyClient(ctx context.Context, obj *unstructured.Unstructured) error {
	// First, try to get the resource, if it exists, update it
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(obj.GroupVersionKind())
	err := r.Get(ctx, client.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)
	if err != nil && !errors.IsNotFound(err) {
		r.Log.Errorf("unable to get existing resource: %v", err)
		return err
	}

	if err == nil {
		// Resource exists, update it
		obj.SetResourceVersion(existing.GetResourceVersion())
		if err := r.Update(ctx, obj); err != nil {
			r.Log.Errorf("unable to update resource: %v", err)
			return err
		}
		r.Log.Info("resource updated successfully")
	} else {
		// Resource does not exist, create it
		if err := r.Create(ctx, obj); err != nil {
			r.Log.Errorf("unable to create resource: %v", err)
			return err
		}
		r.Log.Info("resource created successfully")
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScoringPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionv1beta1.Scoring{}).
		Complete(r)
}
