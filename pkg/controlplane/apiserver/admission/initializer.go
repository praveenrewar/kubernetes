/*
Copyright 2016 The Kubernetes Authors.

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

package admission

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	quota "k8s.io/apiserver/pkg/quota/v1"
)

// PluginInitializer is used for initialization of the generic controlplane admission plugins.
type PluginInitializer struct {
	RESTMapper         meta.RESTMapper
	QuotaConfiguration quota.Configuration
}

var _ admission.PluginInitializer = &PluginInitializer{}

// Initialize checks the initialization interfaces implemented by each plugin
// and provide the appropriate initialization data
func (i *PluginInitializer) Initialize(plugin admission.Interface) {
	if wants, ok := plugin.(initializer.WantsRESTMapper); ok {
		wants.SetRESTMapper(i.RESTMapper)
	}

	if wants, ok := plugin.(initializer.WantsQuotaConfiguration); ok {
		wants.SetQuotaConfiguration(i.QuotaConfiguration)
	}
}
