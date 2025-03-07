/*
Copyright 2021 The Skaffold Authors

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
package v2

import (
	"context"
	"io"
	"time"

	deployutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/instrumentation"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/manifest"
)

// Apply sends Kubernetes manifests to the cluster.
func (r *SkaffoldRunner) Apply(ctx context.Context, out io.Writer) error {
	// TODO: fix empty manifest list and instead read manifests from hydrationDir
	if err := r.applyResources(ctx, out, nil, nil, manifest.ManifestList{}); err != nil {
		return err
	}

	statusCheckOut, postStatusCheckFn, err := deployutil.WithStatusCheckLogFile(time.Now().Format(deployutil.TimeFormat)+".log", out, r.runCtx.Muted())
	postStatusCheckFn()
	if err != nil {
		return err
	}
	sErr := r.deployer.GetStatusMonitor().Check(ctx, statusCheckOut)
	return sErr
}

func (r *SkaffoldRunner) applyResources(ctx context.Context, out io.Writer, artifacts, _ []graph.Artifact, list manifest.ManifestList) error {
	deployOut, postDeployFn, err := deployutil.WithLogFile(time.Now().Format(deployutil.TimeFormat)+".log", out, r.runCtx.Muted())
	if err != nil {
		return err
	}

	event.DeployInProgress()
	ctx, endTrace := instrumentation.StartTrace(ctx, "applyResources_Deploying")
	defer endTrace()
	err = r.deployer.Deploy(ctx, deployOut, artifacts, list)
	postDeployFn()
	if err != nil {
		event.DeployFailed(err)
		endTrace(instrumentation.TraceEndError(err))
		return err
	}
	r.hasDeployed = true
	event.DeployComplete()
	return nil
}
