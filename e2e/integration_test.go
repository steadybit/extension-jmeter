// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"github.com/steadybit/action-kit/go/action_kit_test/client"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-jmeter",
		Port: 8087,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{"--set", "logging.level=debug"}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "run jmeter",
			Test: testRunJMeter,
		},
	})
}

func testRunJMeter(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
		Parameter []map[string]string
	}{
		Parameter: []map[string]string{
			{"key": "Test1", "value": "foo"},
			{"key": "Test2", "value": "bar"},
		},
	}
	files := []client.File{
		{
			ParameterName: "file",
			FileName:      "script.js",
			Content: []byte("" +
				"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<jmeterTestPlan version=\"1.2\" properties=\"5.0\" jmeter=\"5.5\">\n  <hashTree>\n    <TestPlan guiclass=\"TestPlanGui\" testclass=\"TestPlan\" testname=\"Example Plan\" enabled=\"true\">\n      <stringProp name=\"TestPlan.comments\"></stringProp>\n      <boolProp name=\"TestPlan.functional_mode\">false</boolProp>\n      <boolProp name=\"TestPlan.tearDown_on_shutdown\">true</boolProp>\n      <boolProp name=\"TestPlan.serialize_threadgroups\">false</boolProp>\n      <elementProp name=\"TestPlan.user_defined_variables\" elementType=\"Arguments\" guiclass=\"ArgumentsPanel\" testclass=\"Arguments\" testname=\"User Defined Variables\" enabled=\"true\">\n        <collectionProp name=\"Arguments.arguments\"/>\n      </elementProp>\n      <stringProp name=\"TestPlan.user_define_classpath\"></stringProp>\n    </TestPlan>\n    <hashTree>\n      <ThreadGroup guiclass=\"ThreadGroupGui\" testclass=\"ThreadGroup\" testname=\"Example Thread\" enabled=\"true\">\n        <stringProp name=\"ThreadGroup.on_sample_error\">continue</stringProp>\n        <elementProp name=\"ThreadGroup.main_controller\" elementType=\"LoopController\" guiclass=\"LoopControlPanel\" testclass=\"LoopController\" testname=\"Loop Controller\" enabled=\"true\">\n          <boolProp name=\"LoopController.continue_forever\">false</boolProp>\n          <stringProp name=\"LoopController.loops\">1</stringProp>\n        </elementProp>\n        <stringProp name=\"ThreadGroup.num_threads\">1</stringProp>\n        <stringProp name=\"ThreadGroup.ramp_time\">1</stringProp>\n        <boolProp name=\"ThreadGroup.scheduler\">false</boolProp>\n        <stringProp name=\"ThreadGroup.duration\"></stringProp>\n        <stringProp name=\"ThreadGroup.delay\"></stringProp>\n        <boolProp name=\"ThreadGroup.same_user_on_next_iteration\">true</boolProp>\n      </ThreadGroup>\n      <hashTree>\n        <HTTPSamplerProxy guiclass=\"HttpTestSampleGui\" testclass=\"HTTPSamplerProxy\" testname=\"Steadybit Website\" enabled=\"true\">\n          <elementProp name=\"HTTPsampler.Arguments\" elementType=\"Arguments\" guiclass=\"HTTPArgumentsPanel\" testclass=\"Arguments\" testname=\"User Defined Variables\" enabled=\"true\">\n            <collectionProp name=\"Arguments.arguments\"/>\n          </elementProp>\n          <stringProp name=\"HTTPSampler.domain\">www.steadybit.com</stringProp>\n          <stringProp name=\"HTTPSampler.port\"></stringProp>\n          <stringProp name=\"HTTPSampler.protocol\">https</stringProp>\n          <stringProp name=\"HTTPSampler.contentEncoding\"></stringProp>\n          <stringProp name=\"HTTPSampler.path\"></stringProp>\n          <stringProp name=\"HTTPSampler.method\">GET</stringProp>\n          <boolProp name=\"HTTPSampler.follow_redirects\">true</boolProp>\n          <boolProp name=\"HTTPSampler.auto_redirects\">false</boolProp>\n          <boolProp name=\"HTTPSampler.use_keepalive\">true</boolProp>\n          <boolProp name=\"HTTPSampler.DO_MULTIPART_POST\">false</boolProp>\n          <stringProp name=\"HTTPSampler.embedded_url_re\"></stringProp>\n          <stringProp name=\"HTTPSampler.connect_timeout\"></stringProp>\n          <stringProp name=\"HTTPSampler.response_timeout\"></stringProp>\n        </HTTPSamplerProxy>\n        <hashTree/>\n      </hashTree>\n    </hashTree>\n  </hashTree>\n</jmeterTestPlan>"),
		},
	}
	exec, err := e.RunActionWithFiles("com.steadybit.extension_jmeter.run", nil, config, nil, files)
	require.NoError(t, err)
	e2e.AssertProcessRunningInContainer(t, m, e.Pod, "extension", "jmeter", true)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Waiting for possible Shutdown/StopTestNow/HeapDump/ThreadDump message on port 4445", 60*time.Second)
	require.NoError(t, exec.Cancel())
}
