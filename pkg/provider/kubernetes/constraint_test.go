/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

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

package kubernetes_test

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"

	kubermaticv1 "k8c.io/kubermatic/v2/pkg/crd/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/provider/kubernetes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/kubernetes/scheme"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testClusterName = "test-constraints"
	testNamespace   = "cluster-test-constraints"
)

func TestListConstraints(t *testing.T) {

	testCases := []struct {
		name                string
		existingObjects     []runtime.Object
		cluster             *kubermaticv1.Cluster
		expectedConstraints []*kubermaticv1.Constraint
	}{
		{
			name: "scenario 1: list constraints",
			existingObjects: []runtime.Object{
				genConstraint("ct1", testNamespace),
				genConstraint("ct2", testNamespace),
				genConstraint("ct3", "other-ns"),
			},
			cluster:             genCluster(testClusterName, "kubernetes", "my-first-project-ID", "test-constraints", "john@acme.com"),
			expectedConstraints: []*kubermaticv1.Constraint{genConstraint("ct1", testNamespace), genConstraint("ct2", testNamespace)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fakectrlruntimeclient.NewFakeClientWithScheme(scheme.Scheme, tc.existingObjects...)
			constraintProvider, err := kubernetes.NewConstraintProvider(client)
			if err != nil {
				t.Fatal(err)
			}

			constraintList, err := constraintProvider.List(tc.cluster)
			if err != nil {
				t.Fatal(err)
			}
			if len(tc.expectedConstraints) != len(constraintList.Items) {
				t.Fatalf("expected to get %d constraints, but got %d", len(tc.expectedConstraints), len(constraintList.Items))
			}

			for _, returnedConstraint := range constraintList.Items {
				cFound := false
				for _, expectedCT := range tc.expectedConstraints {
					if dif := deep.Equal(returnedConstraint, *expectedCT); dif == nil {
						cFound = true
						break
					}
				}
				if !cFound {
					t.Fatalf("returned constraint was not found on the list of expected ones, ct = %#v", returnedConstraint)
				}
			}
		})
	}
}

func TestGetConstraint(t *testing.T) {

	testCases := []struct {
		name               string
		existingObjects    []runtime.Object
		cluster            *kubermaticv1.Cluster
		expectedConstraint *kubermaticv1.Constraint
	}{
		{
			name: "scenario 1: get constraint",
			existingObjects: []runtime.Object{
				genConstraint("ct1", testNamespace),
				genConstraint("ct2", testNamespace),
				genConstraint("ct3", "other-ns"),
			},
			cluster:            genCluster(testClusterName, "kubernetes", "my-first-project-ID", "test-constraints", "john@acme.com"),
			expectedConstraint: genConstraint("ct1", testNamespace),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := fakectrlruntimeclient.NewFakeClientWithScheme(scheme.Scheme, tc.existingObjects...)
			constraintProvider, err := kubernetes.NewConstraintProvider(client)
			if err != nil {
				t.Fatal(err)
			}

			constraint, err := constraintProvider.Get(tc.cluster, tc.expectedConstraint.Name)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(constraint, tc.expectedConstraint) {
				t.Fatalf(" diff: %s", diff.ObjectGoPrintSideBySide(constraint, tc.expectedConstraint))
			}
		})
	}
}

func genConstraint(name, namespace string) *kubermaticv1.Constraint {
	ct := &kubermaticv1.Constraint{}
	ct.Kind = kubermaticv1.ConstraintKind
	ct.APIVersion = kubermaticv1.SchemeGroupVersion.String()
	ct.Name = name
	ct.Namespace = namespace
	ct.Spec = kubermaticv1.ConstraintSpec{
		ConstraintType: "requiredlabels",
		Match: kubermaticv1.Match{
			Kinds: []kubermaticv1.Kind{
				{Kinds: "namespace", APIGroups: ""},
			},
		},
		Parameters: kubermaticv1.Parameters{
			RawJSON: `{"labels":[ "gatekeeper", "opa"]}`,
		},
	}

	return ct
}
