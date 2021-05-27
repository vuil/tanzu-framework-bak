// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kind_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/kind"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

var (
	testingDir                            string
	defaultBoMFileForTesting              = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	configPath                            = "../fakes/config/config6.yaml"
	configPathCustomRegistrySkipTLSVerify = "../fakes/config/config_custom_registry_skip_tls_verify.yaml"
	configPathCustomRegistryCaCert        = "../fakes/config/config_custom_registry_ca_cert.yaml"
	registryHostname                      = "registry.mydomain.com"
)

func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kind Suite")
}

var (
	kindClient   kind.Client
	kindProvider *fakes.KindProvider
	err          error
	clusterName  string
	kindConfig   string
)

var _ = Describe("Kind Client", func() {
	BeforeSuite(func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	})

	AfterSuite(func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	})

	Context("When TKG_CUSTOM_IMAGE_REPOSITORY is not set", func() {
		BeforeEach(func() {
			setupTestingFiles(configPath, testingDir, defaultBoMFileForTesting)
			kindClient = buildKindClient()
		})

		Describe("Create bootstrap kind cluster", func() {
			JustBeforeEach(func() {
				clusterName, err = kindClient.CreateKindCluster()
			})

			Context("When kind provider fails to create kind cluster", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("failed to create kind cluster"))
				})
			})

			Context("When kind provider create kind cluster but unable to retrieve kubeconfig", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(nil)
					kindProvider.KubeConfigReturns("", errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("unable to retrieve kubeconfig for created kind cluster"))
				})
			})

			Context("When kind provider create kind cluster and able to retrieve kubeconfig successfully", func() {
				BeforeEach(func() {
					kindProvider.CreateReturns(nil)
					kindProvider.KubeConfigReturns("fake-kube-config", nil)
				})
				It("does not return error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(clusterName).To(Equal("clusterName"))
				})
			})
		})

		Describe("Delete bootstrap kind cluster", func() {
			JustBeforeEach(func() {
				err = kindClient.DeleteKindCluster()
			})

			Context("When kind provider fails to delete kind cluster", func() {
				BeforeEach(func() {
					kindProvider.DeleteReturns(errors.New("fake-error"))
				})
				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("failed to delete kind cluster"))
				})
			})

			Context("When kind provider deletes kind cluster successfully", func() {
				BeforeEach(func() {
					kindProvider.DeleteReturns(nil)
				})
				It("returns an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Context("When TKG_CUSTOM_IMAGE_REPOSITORY is set", func() {
		Context("When TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY is set to true", func() {
			BeforeEach(func() {
				setupTestingFiles(configPathCustomRegistrySkipTLSVerify, testingDir, defaultBoMFileForTesting)
				kindClient = buildKindClient()
				_, kindConfigBytes, err := kindClient.GetKindNodeImageAndConfig()
				Expect(err).NotTo(HaveOccurred())
				kindConfig = string(kindConfigBytes)
			})

			Describe("Generate kind cluster config", func() {
				It("generates 'insecure_skip_verify = true' in containerdConfigPatches", func() {
					Expect(kindConfig).Should(ContainSubstring(fmt.Sprintf(kind.KindRegistryConfigSkipTLSVerify, registryHostname)))
					Expect(kindConfig).Should(ContainSubstring("insecure_skip_verify = true"))
					Expect(kindConfig).ShouldNot(ContainSubstring("ca_file = \"/etc/containerd/tkg-registry-ca.crt\""))
					Expect(kindConfig).ShouldNot(ContainSubstring("containerPath: /etc/containerd/tkg-registry-ca.crt"))
				})
			})
		})

		Context("When TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE is set to non-empty string", func() {
			BeforeEach(func() {
				setupTestingFiles(configPathCustomRegistryCaCert, testingDir, defaultBoMFileForTesting)
				kindClient = buildKindClient()
				_, kindConfigBytes, err := kindClient.GetKindNodeImageAndConfig()
				Expect(err).NotTo(HaveOccurred())
				kindConfig = string(kindConfigBytes)
			})

			Describe("Generate kind cluster config", func() {
				It("generates ca_file config in containerdConfigPatches", func() {
					Expect(kindConfig).Should(ContainSubstring(fmt.Sprintf(kind.KindRegistryConfigCaCert, registryHostname)))
					Expect(kindConfig).Should(ContainSubstring("insecure_skip_verify = false"))
					Expect(kindConfig).Should(ContainSubstring("ca_file = \"/etc/containerd/tkg-registry-ca.crt\""))
					Expect(kindConfig).Should(ContainSubstring("containerPath: /etc/containerd/tkg-registry-ca.crt"))
				})
			})
		})
	})
})

func buildKindClient() kind.Client {
	tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
	Expect(err).NotTo(HaveOccurred())

	kindProvider = &fakes.KindProvider{}
	options := kind.KindClusterOptions{
		Provider:       kindProvider,
		ClusterName:    "clusterName",
		NodeImage:      "nodeImage",
		KubeConfigPath: "kubeConfigPath",
		TKGConfigDir:   testingDir,
		Readerwriter:   tkgConfigReaderWriter,
	}
	return kind.New(&options)
}

func setupTestingFiles(clusterConfigFile string, configDir string, defaultBomFile string) {
	testClusterConfigFile := filepath.Join(configDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	tkgconfigpaths.TKGDefaultBOMImageTag = utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())
}