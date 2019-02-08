// Copyright 2019 The Google Cloud Robotics Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package approllout

import (
	"strings"
	"testing"

	apps "github.com/googlecloudrobotics/core/src/go/pkg/apis/apps/v1alpha1"
	registry "github.com/googlecloudrobotics/core/src/go/pkg/apis/registry/v1alpha1"
	"k8s.io/helm/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

func marshalYAML(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func unmarshalYAML(t *testing.T, v interface{}, s string) {
	t.Helper()
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(s)), v); err != nil {
		t.Fatal(err)
	}
}

func TestNewRobotChartAssignment(t *testing.T) {
	var app apps.App
	unmarshalYAML(t, &app, `
metadata:
  name: foo
spec:
  repository: https://example.org/helm
  version: 1.2.3
  components:
    robot:
      name: foo-robot
      inline: abcdefgh
	`)

	var rollout apps.AppRollout
	unmarshalYAML(t, &rollout, `
metadata:
  name: foo-rollout
  labels:
    lkey1: lval1
  annotations:
    akey1: aval1
spec:
  appName: prometheus
  robots:
  - selector:
      any: true
    values:
      foo1: bar1
    version: 1.2.4
 `)

	var robot registry.Robot
	unmarshalYAML(t, &robot, `
metadata:
  name: robot1
	`)

	baseValues := chartutil.Values{
		"foo2": "bar2",
	}

	var expected apps.ChartAssignment
	unmarshalYAML(t, &expected, `
metadata:
  name: foo-rollout-robot.robot1
  labels:
    lkey1: lval1
  annotations:
    akey1: aval1
spec:
  clusterName: robot1
  namespaceName: app-foo-rollout
  chart:
    repository: https://example.org/helm
    version: 1.2.4
    name: foo-robot
    inline: abcdefgh
    values:
      foo1: bar1
      foo2: bar2
	`)

	result := newRobotChartAssignment(&robot, &app, &rollout, &rollout.Spec.Robots[0], baseValues)
	// Compare serialized YAML for easier diff detection and to avoid complicated
	// comparisons for map[string]interface{} values.
	expectedStr := marshalYAML(t, expected)
	resultStr := marshalYAML(t, result)

	if expectedStr != resultStr {
		t.Fatalf("expected ChartAssignment: \n%s\ngot:\n%s\n", expectedStr, resultStr)
	}
}

func TestNewCloudChartAssignment(t *testing.T) {
	var app apps.App
	unmarshalYAML(t, &app, `
metadata:
  name: foo
spec:
  repository: https://example.org/helm
  version: 1.2.3
  components:
    cloud:
      name: foo-cloud
      inline: abcdefgh
	`)

	var rollout apps.AppRollout
	unmarshalYAML(t, &rollout, `
metadata:
  name: foo-rollout
  labels:
    lkey1: lval1
  annotations:
    akey1: aval1
spec:
  appName: prometheus
  cloud:
    values:
      robots: should_be_overwritten
      foo1: bar1
 `)

	var robot1, robot2 registry.Robot
	unmarshalYAML(t, &robot1, `
metadata:
  name: robot1
	`)
	unmarshalYAML(t, &robot2, `
metadata:
  name: robot2
	`)

	baseValues := chartutil.Values{
		"foo2": "bar2",
	}

	var expected apps.ChartAssignment
	unmarshalYAML(t, &expected, `
metadata:
  name: foo-rollout-cloud
  labels:
    lkey1: lval1
  annotations:
    akey1: aval1
spec:
  clusterName: cloud
  namespaceName: app-foo-rollout
  chart:
    repository: https://example.org/helm
    version: 1.2.3
    name: foo-cloud
    inline: abcdefgh
    values:
      robots:
      - name: robot1
      - name: robot2
      foo1: bar1
      foo2: bar2
	`)

	result := newCloudChartAssignment(&app, &rollout, baseValues, &robot1, &robot2)
	// Compare serialized YAML for easier diff detection and to avoid complicated
	// comparisons for map[string]interface{} values.
	expectedStr := marshalYAML(t, expected)
	resultStr := marshalYAML(t, result)

	if expectedStr != resultStr {
		t.Fatalf("expected ChartAssignment: \n%s\ngot:\n%s\n", expectedStr, resultStr)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name       string
		cur        string
		shouldFail bool
	}{
		{
			name: "valid-all",
			cur: `
spec:
  appName: myapp
  cloud:
    values:
      a: 2
      b: {c: 3}
  robots:
  - selector:
      any: true
    values:
      c: d
  - selector:
      matchLabels:
        abc: def
        foo: bar
  - selector:
      matchExpressions:
      - {key: foo, Op: DoesExist}
	`,
		},
		{
			name: "valid-app-name-only",
			cur: `
spec:
  appName: my-app.123
	`,
		},
		{
			name:       "missing-app-name",
			cur:        `spec: {}`,
			shouldFail: true,
		},
		{
			name: "invalid-app-name",
			cur: `
spec:
  appName: my%app
	`,
			shouldFail: true,
		},
		{
			name: "missing-robot-selector",
			cur: `
spec:
  appName: myapp
  robots:
  - values:
      a: b
	`,
			shouldFail: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var cur apps.AppRollout
			unmarshalYAML(t, &cur, c.cur)

			err := validate(&cur)
			if err == nil && c.shouldFail {
				t.Fatal("expected failure but got none")
			}
			if err != nil && !c.shouldFail {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
