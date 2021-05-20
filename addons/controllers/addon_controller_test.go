// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/vmware-tanzu-private/core/addons/testutil"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/cluster-api/util/secret"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonconstants "github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
)

const (
	waitTimeout     = time.Second * 60
	pollingInterval = time.Second * 1
)

var _ = Describe("Addon Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster, tkr, BOM config map and addon secret")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, "default", k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting cluster, tkr, BOM config map and addon secret")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		Expect(testutil.DeleteResources(f, cfg, dynamicClient, true)).To(Succeed())

		By("Deleting Addon data-values secrets")
		addonSecretKey := client.ObjectKey{
			Namespace: addonconstants.TKGAddonsAppNamespace,
			Name:      "antrea-data-values",
		}
		dataValuesSecret := &v1.Secret{}
		Expect(k8sClient.Get(ctx, addonSecretKey, dataValuesSecret)).To(Succeed())
		Expect(k8sClient.Delete(ctx, dataValuesSecret)).To(Succeed())

		By("Deleting Addon app CR")
		appKey := client.ObjectKey{
			Namespace: addonconstants.TKGAddonsAppNamespace,
			Name:      "antrea",
		}
		antreaApp := &kappctrl.App{}
		Expect(k8sClient.Get(ctx, appKey, antreaApp)).To(Succeed())
		Expect(k8sClient.Delete(ctx, antreaApp)).To(Succeed())

		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: "default",
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &v1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	Context("reconcileAddonNormal for a tkr 1.18.1", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-1"
			clusterResourceFilePath = "testdata/test-cluster-1.yaml"
		})

		It("Should create addon namespace, service account cluster admin service role and role binding", func() {

			Eventually(func() bool {
				ns := &v1.NamespaceList{}
				err := k8sClient.List(ctx, ns)
				if err != nil {
					return false
				}
				for _, n := range ns.Items {
					if n.Name == addonconstants.TKGAddonsAppNamespace {
						return true
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonconstants.TKGAddonsAppNamespace,
					Name:      addonconstants.TKGAddonsAppServiceAccount,
				}
				svc := &v1.ServiceAccount{}
				err := k8sClient.Get(ctx, key, svc)
				return err == nil
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				roles := &rbacv1.ClusterRoleList{}
				err := k8sClient.List(ctx, roles)
				if err != nil {
					return false
				}
				for _, r := range roles.Items {
					if r.Name == addonconstants.TKGAddonsAppClusterRole {
						rule := r.Rules[0]
						if rule.APIGroups[0] == "*" && rule.Verbs[0] == "*" && rule.Resources[0] == "*" {
							return true
						}
					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				roleBindings := &rbacv1.ClusterRoleBindingList{}
				err := k8sClient.List(ctx, roleBindings)
				if err != nil {
					return false
				}
				for _, r := range roleBindings.Items {
					if r.Name == addonconstants.TKGAddonsAppClusterRoleBinding &&
						r.RoleRef.Name == addonconstants.TKGAddonsAppClusterRole {
						if r.Subjects[0].Name == addonconstants.TKGAddonsAppServiceAccount &&
							r.Subjects[0].Namespace == addonconstants.TKGAddonsAppNamespace {
							return true
						}

					}
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

		It("Addon controller reconciliation check", func() {

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonconstants.TKGAddonsAppNamespace,
					Name:      "antrea-data-values",
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, key, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "serviceCidr: 100.64.0.0/13")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonconstants.TKGAddonsAppNamespace,
					Name:      "antrea",
				}
				app := &kappctrl.App{}
				Expect(k8sClient.Get(ctx, key, app)).To(Succeed())

				Expect(app.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("cni/antrea"))
				Expect(app.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-1-antrea"))
				// TODO why is this needed
				Expect(app.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				Expect(app.Spec.ServiceAccountName).Should(Equal(addonconstants.TKGAddonsAppServiceAccount))

				Expect(app.Spec.Fetch[0].Image.URL).Should(Equal("projects-stg.registry.vmware.com/tkg/addons/antrea-templates:98adbf4"))

				appTmplYtt := kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									Name: "antrea-data-values",
								},
							},
						},
					},
				}

				Expect(*app.Spec.Template[0].Ytt).Should(Equal(appTmplYtt))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})
	})

	Context("reconcileAddonNormal for a tkr 1.20.5", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-2"
			clusterResourceFilePath = "testdata/test-cluster-2.yaml"
		})

		It("Addon controller reconciliation check", func() {

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonconstants.TKGAddonsAppNamespace,
					Name:      "antrea-data-values",
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, key, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "serviceCidr: 100.64.0.0/13")).Should(BeTrue())
				imageInfoData := string(secret.Data["imageInfo.yaml"])
				Expect(strings.Contains(imageInfoData, "imageRepository: projects.registry.vmware.com/tkg")).Should(BeTrue())
				Expect(strings.Contains(imageInfoData, "imagePath: antrea/antrea-debian")).Should(BeTrue())
				Expect(strings.Contains(imageInfoData, "tag: v0.11.3_vmware.2")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				key := client.ObjectKey{
					Namespace: addonconstants.TKGAddonsAppNamespace,
					Name:      "antrea",
				}
				app := &kappctrl.App{}
				Expect(k8sClient.Get(ctx, key, app)).To(Succeed())

				Expect(app.Annotations[addontypes.AddonTypeAnnotation]).Should(Equal("cni/antrea"))
				Expect(app.Annotations[addontypes.AddonNameAnnotation]).Should(Equal("test-cluster-2-antrea"))
				// TODO why is this needed
				Expect(app.Annotations[addontypes.AddonNamespaceAnnotation]).Should(Equal("default"))

				Expect(app.Spec.ServiceAccountName).Should(Equal(addonconstants.TKGAddonsAppServiceAccount))

				Expect(app.Spec.Fetch[0].Image.URL).Should(Equal("projects.registry.vmware.com/tkg/tanzu_core/addons/antrea-templates:v1.3.1"))

				appTmplYtt := kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "antrea-data-values",
									},
								},
							},
						},
					},
				}

				Expect(*app.Spec.Template[0].Ytt).Should(Equal(appTmplYtt))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

})
