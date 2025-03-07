/*
Copyright 2022 The Skaffold Authors

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

package kubectl

import (
	"context"
	"io"

	apimachinery "k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/instrumentation"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/manifest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/output/log"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/render"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/render/generate"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/render/renderer/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

type Kubectl struct {
	generate.Generator
	labels map[string]string

	transformAllowlist map[apimachinery.GroupKind]latest.ResourceFilter
	transformDenylist  map[apimachinery.GroupKind]latest.ResourceFilter
}

func New(cfg render.Config, labels map[string]string) (Kubectl, error) {
	generator := generate.NewGenerator(cfg.GetWorkingDir(), cfg.GetRenderConfig().Generate)
	transformAllowlist, transformDenylist, err := util.ConsolidateTransformConfiguration(cfg)
	if err != nil {
		return Kubectl{}, err
	}
	return Kubectl{
		Generator: generator,
		labels:    labels,

		transformAllowlist: transformAllowlist,
		transformDenylist:  transformDenylist,
	}, nil
}

func (r Kubectl) Render(ctx context.Context, out io.Writer, builds []graph.Artifact, _ bool, _ string) (manifest.ManifestList, error) {
	_, endTrace := instrumentation.StartTrace(ctx, "Render_KubectlManifests")
	log.Entry(ctx).Infof("rendering using kubectl")
	instrumentation.AddAttributesToCurrentSpanFromContext(ctx, map[string]string{
		"RendererType": "kubectl",
	})
	manifests, err := util.GenerateHydratedManifests(ctx, out, builds, r.Generator, r.labels, r.transformAllowlist, r.transformDenylist)
	endTrace()
	return manifests, err
}
