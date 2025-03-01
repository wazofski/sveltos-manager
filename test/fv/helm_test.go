/*
Copyright 2022. projectsveltos.io. All rights reserved.

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

package fv_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	configv1alpha1 "github.com/projectsveltos/sveltos-manager/api/v1alpha1"
)

var _ = Describe("Helm", func() {
	const (
		namePrefix = "helm-"
	)

	It("Deploy and updates helm charts correctly", Label("FV"), func() {
		Byf("Create a ClusterProfile matching Cluster %s/%s", kindWorkloadCluster.Namespace, kindWorkloadCluster.Name)
		clusterProfile := getClusterProfile(namePrefix, map[string]string{key: value})
		clusterProfile.Spec.SyncMode = configv1alpha1.SyncModeContinuous
		Expect(k8sClient.Create(context.TODO(), clusterProfile)).To(Succeed())

		verifyClusterProfileMatches(clusterProfile)

		verifyClusterSummary(clusterProfile, kindWorkloadCluster.Namespace, kindWorkloadCluster.Name)

		Byf("Update ClusterProfile %s to deploy helm charts", clusterProfile.Name)
		currentClusterProfile := &configv1alpha1.ClusterProfile{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: clusterProfile.Name}, currentClusterProfile)).To(Succeed())
		currentClusterProfile.Spec.HelmCharts = []configv1alpha1.HelmChart{
			{
				RepositoryURL:    "https://kyverno.github.io/kyverno/",
				RepositoryName:   "kyverno",
				ChartName:        "kyverno/kyverno",
				ChartVersion:     "v2.6.5",
				ReleaseName:      "kyverno-latest",
				ReleaseNamespace: "kyverno",
				HelmChartAction:  configv1alpha1.HelmChartActionInstall,
			},
			{
				RepositoryURL:    "https://helm.nginx.com/stable/",
				RepositoryName:   "nginx-stable",
				ChartName:        "nginx-stable/nginx-ingress",
				ChartVersion:     "0.14.0",
				ReleaseName:      "nginx-latest",
				ReleaseNamespace: "nginx",
				HelmChartAction:  configv1alpha1.HelmChartActionInstall,
			},
		}
		Expect(k8sClient.Update(context.TODO(), currentClusterProfile)).To(Succeed())

		clusterSummary := verifyClusterSummary(currentClusterProfile, kindWorkloadCluster.Namespace, kindWorkloadCluster.Name)

		Byf("Getting client to access the workload cluster")
		workloadClient, err := getKindWorkloadClusterKubeconfig()
		Expect(err).To(BeNil())
		Expect(workloadClient).ToNot(BeNil())

		Byf("Verifying kyverno deployment is created in the workload cluster")
		Eventually(func() error {
			depl := &appsv1.Deployment{}
			return workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "kyverno", Name: "kyverno-latest"}, depl)
		}, timeout, pollingInterval).Should(BeNil())

		Byf("Verifying nginx deployment is created in the workload cluster")
		Eventually(func() error {
			depl := &appsv1.Deployment{}
			return workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "nginx", Name: "nginx-latest-nginx-ingress"}, depl)
		}, timeout, pollingInterval).Should(BeNil())

		Byf("Verifying ClusterSummary %s status is set to Deployed for Helm feature", clusterSummary.Name)
		verifyFeatureStatusIsProvisioned(kindWorkloadCluster.Namespace, clusterSummary.Name, configv1alpha1.FeatureHelm)

		charts := []configv1alpha1.Chart{
			{ReleaseName: "kyverno-latest", ChartVersion: "2.6.5", Namespace: "kyverno"},
			{ReleaseName: "nginx-latest", ChartVersion: "0.14.0", Namespace: "nginx"},
		}

		verifyClusterConfiguration(clusterProfile.Name, clusterSummary.Spec.ClusterNamespace,
			clusterSummary.Spec.ClusterName, configv1alpha1.FeatureHelm, nil, charts)

		Byf("Changing ClusterProfile requiring different chart version for kyverno")
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: clusterProfile.Name}, currentClusterProfile)).To(Succeed())
		currentClusterProfile.Spec.HelmCharts = []configv1alpha1.HelmChart{
			{
				RepositoryURL:    "https://kyverno.github.io/kyverno/",
				RepositoryName:   "kyverno",
				ChartName:        "kyverno/kyverno",
				ChartVersion:     "v2.6.4",
				ReleaseName:      "kyverno-latest",
				ReleaseNamespace: "kyverno",
				HelmChartAction:  configv1alpha1.HelmChartActionInstall,
			},
			{
				RepositoryURL:    "https://helm.nginx.com/stable/",
				RepositoryName:   "nginx-stable",
				ChartName:        "nginx-stable/nginx-ingress",
				ChartVersion:     "0.14.0",
				ReleaseName:      "nginx-latest",
				ReleaseNamespace: "nginx",
				HelmChartAction:  configv1alpha1.HelmChartActionInstall,
			},
		}
		Expect(k8sClient.Update(context.TODO(), currentClusterProfile)).To(Succeed())

		verifyClusterSummary(currentClusterProfile, kindWorkloadCluster.Namespace, kindWorkloadCluster.Name)

		Byf("Verifying kyverno deployment is still present in the workload cluster")
		Eventually(func() error {
			depl := &appsv1.Deployment{}
			return workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "kyverno", Name: "kyverno-latest"}, depl)
		}, timeout, pollingInterval).Should(BeNil())

		Byf("Verifying nginx deployment is still in the workload cluster")
		Eventually(func() error {
			depl := &appsv1.Deployment{}
			return workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "nginx", Name: "nginx-latest-nginx-ingress"}, depl)
		}, timeout, pollingInterval).Should(BeNil())

		Byf("Verifying ClusterSummary %s status is set to Deployed for Helm feature", clusterSummary.Name)
		verifyFeatureStatusIsProvisioned(kindWorkloadCluster.Namespace, clusterSummary.Name, configv1alpha1.FeatureHelm)

		charts = []configv1alpha1.Chart{
			{ReleaseName: "kyverno-latest", ChartVersion: "2.6.4", Namespace: "kyverno"},
			{ReleaseName: "nginx-latest", ChartVersion: "0.14.0", Namespace: "nginx"},
		}

		verifyClusterConfiguration(clusterProfile.Name, clusterSummary.Spec.ClusterNamespace,
			clusterSummary.Spec.ClusterName, configv1alpha1.FeatureHelm, nil, charts)

		Byf("Changing ClusterProfile to not require nginx anymore")
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: clusterProfile.Name}, currentClusterProfile)).To(Succeed())
		currentClusterProfile.Spec.HelmCharts = []configv1alpha1.HelmChart{
			{
				RepositoryURL:    "https://kyverno.github.io/kyverno/",
				RepositoryName:   "kyverno",
				ChartName:        "kyverno/kyverno",
				ChartVersion:     "v2.6.4",
				ReleaseName:      "kyverno-latest",
				ReleaseNamespace: "kyverno",
				HelmChartAction:  configv1alpha1.HelmChartActionInstall,
			},
		}
		Expect(k8sClient.Update(context.TODO(), currentClusterProfile)).To(Succeed())

		verifyClusterSummary(currentClusterProfile, kindWorkloadCluster.Namespace, kindWorkloadCluster.Name)

		Byf("Verifying kyverno deployment is still present in the workload cluster")
		Eventually(func() error {
			depl := &appsv1.Deployment{}
			return workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "kyverno", Name: "kyverno-latest"}, depl)
		}, timeout, pollingInterval).Should(BeNil())

		Byf("Verifying nginx deployment is removed from workload cluster")
		Eventually(func() bool {
			depl := &appsv1.Deployment{}
			err = workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "nginx", Name: "nginx-latest-nginx-ingress"}, depl)
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		charts = []configv1alpha1.Chart{
			{ReleaseName: "kyverno-latest", ChartVersion: "2.6.4", Namespace: "kyverno"},
		}

		verifyClusterConfiguration(clusterProfile.Name, clusterSummary.Spec.ClusterNamespace,
			clusterSummary.Spec.ClusterName, configv1alpha1.FeatureHelm, nil, charts)

		deleteClusterProfile(clusterProfile)

		Byf("Verifying kyverno deployment is removed from workload cluster")
		Eventually(func() bool {
			depl := &appsv1.Deployment{}
			err = workloadClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "kyverno", Name: "kyverno-latest"}, depl)
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())
	})
})
