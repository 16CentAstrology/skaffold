/*
Copyright 2020 The Skaffold Authors

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

package integration

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/GoogleContainerTools/skaffold/v2/integration/skaffold"
)

func TestModules_BuildDependency(t *testing.T) {
	t.Run("build dependency between artifacts of different modules", func(t *testing.T) {
		MarkIntegrationTest(t, CanRunWithoutGcp)
		ns, client := SetupNamespace(t)

		skaffold.Dev().InDir("testdata/modules").InNs(ns.Name).RunBackground(t)
		client.waitForDeploymentsToStabilizeWithTimeout(3*time.Minute, "app1", "app2", "app3")

		dep1 := client.GetDeployment("app1")
		dep2 := client.GetDeployment("app2")
		dep3 := client.GetDeployment("app3")

		// Make a change to app3/foo so that dev is forced to delete the Deployment and redeploy app1, app2 and app3,
		// since app2 depends on app3 and app1 depends on app2
		Run(t, "testdata/modules/app3", "sh", "-c", "echo bar > foo")
		defer Run(t, "testdata/modules/app3", "sh", "-c", "> foo")
		client.waitForDeploymentsToStabilizeWithTimeout(3*time.Minute, "app1", "app2", "app3")

		// Make sure the old Deployment and the new Deployment are different
		err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
			newDep1 := client.GetDeployment("app1")
			newDep2 := client.GetDeployment("app2")
			newDep3 := client.GetDeployment("app3")
			t.Logf("app1 - old gen: %d, new gen: %d", dep1.GetGeneration(), newDep1.GetGeneration())
			t.Logf("app2 - old gen: %d, new gen: %d", dep2.GetGeneration(), newDep2.GetGeneration())
			t.Logf("app3 - old gen: %d, new gen: %d", dep3.GetGeneration(), newDep3.GetGeneration())
			return dep1.GetGeneration() != newDep1.GetGeneration() &&
				dep2.GetGeneration() != newDep2.GetGeneration() &&
				dep3.GetGeneration() != newDep3.GetGeneration(), nil
		})
		failNowIfError(t, err)
	})
}
