// +build knative

// To enable compilation of this file in Goland, go to "Settings -> Go -> Vendoring & Build Tags -> Custom Tags" and add "knative"

/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

func TestRunServiceCombo(t *testing.T) {
	withNewTestNamespace(t, func(ns string) {

		Expect(kamel("install", "-n", ns, "--trait-profile", "knative").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knative2.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knative2"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Expect(kamel("run", "-n", ns, "files/knative3.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knative3"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Expect(kamel("run", "-n", ns, "files/knative1.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knative1"), testTimeoutLong).Should(Equal(v1.PodRunning))
		// Correct logs
		Eventually(integrationLogs(ns, "knative1"), testTimeoutMedium).Should(ContainSubstring("Received from 2: Hello from knative2"))
		Eventually(integrationLogs(ns, "knative1"), testTimeoutMedium).Should(ContainSubstring("Received from 3: Hello from knative3"))
		// Incorrect logs
		Consistently(integrationLogs(ns, "knative1"), 10*time.Second).ShouldNot(ContainSubstring("Received from 2: Hello from knative3"))
		Consistently(integrationLogs(ns, "knative1"), 10*time.Second).ShouldNot(ContainSubstring("Received from 3: Hello from knative2"))
		// Cleanup
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}

func TestRunChannelCombo(t *testing.T) {
	withNewTestNamespace(t, func(ns string) {

		Expect(createKnativeChannel(ns, "messages")()).Should(BeNil())
		Expect(kamel("install", "-n", ns, "--trait-profile", "knative").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativech2.groovy").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativech1.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knativech2"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationPodPhase(ns, "knativech1"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationLogs(ns, "knativech2"), testTimeoutMedium).Should(ContainSubstring("Received: Hello from knativech1"))
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}

func TestRunChannelComboGetToPost(t *testing.T) {
	withNewTestNamespace(t, func(ns string) {

		Expect(createKnativeChannel(ns, "messages")()).Should(BeNil())
		Expect(kamel("install", "-n", ns, "--trait-profile", "knative").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativegetpost2.groovy").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativegetpost1.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knativegetpost2"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationPodPhase(ns, "knativegetpost1"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationLogs(ns, "knativegetpost2"), testTimeoutMedium).Should(ContainSubstring(`Received ""`))
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}

/*
// FIXME: uncomment when https://github.com/apache/camel-k-runtime/issues/69 is resolved
func TestRunMultiChannelChain(t *testing.T) {
	withNewTestNamespace(t, func(ns string) {
		Expect(createKnativeChannel(ns, "messages")()).Should(BeNil())
		Expect(createKnativeChannel(ns, "words")()).Should(BeNil())
		Expect(kamel("install", "-n", ns, "--trait-profile", "knative").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativemultihop3.groovy").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativemultihop2.groovy").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativemultihop1.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knativemultihop3"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationPodPhase(ns, "knativemultihop2"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationPodPhase(ns, "knativemultihop1"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationLogs(ns, "knativemultihop3"), testTimeoutMedium).Should(ContainSubstring(`From messages: message`))
		Eventually(integrationLogs(ns, "knativemultihop3"), testTimeoutMedium).Should(ContainSubstring(`From words: word`))
		Eventually(integrationLogs(ns, "knativemultihop3"), testTimeoutMedium).Should(ContainSubstring(`From words: transformed message`))
		Eventually(integrationLogs(ns, "knativemultihop3"), 10*time.Second).ShouldNot(ContainSubstring(`From messages: word`))
		Eventually(integrationLogs(ns, "knativemultihop3"), 10*time.Second).ShouldNot(ContainSubstring(`From words: message`))
		Eventually(integrationLogs(ns, "knativemultihop3"), 10*time.Second).ShouldNot(ContainSubstring(`From messages: transformed message`))
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}
*/

func TestRunBroker(t *testing.T) {
	withNewTestNamespaceWithKnativeBroker(t, func(ns string) {
		Expect(kamel("install", "-n", ns, "--trait-profile", "knative").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativeevt1.groovy").Execute()).Should(BeNil())
		Expect(kamel("run", "-n", ns, "files/knativeevt2.groovy").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "knativeevt1"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationPodPhase(ns, "knativeevt2"), testTimeoutLong).Should(Equal(v1.PodRunning))
		Eventually(integrationLogs(ns, "knativeevt2"), testTimeoutMedium).Should(ContainSubstring("Received 1: Hello 1"))
		Eventually(integrationLogs(ns, "knativeevt2"), testTimeoutMedium).Should(ContainSubstring("Received 2: Hello 2"))
		Eventually(integrationLogs(ns, "knativeevt2")).ShouldNot(ContainSubstring("Received 1: Hello 2"))
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}
