// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	fakehelper "github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

const (
	constConfigPath  = "../fakes/config/config.yaml"
	constConfig2Path = "../fakes/config/config2.yaml"
	constConfig3Path = "../fakes/config/config3.yaml"
	constKeyFOO      = "FOO"
	constKeyBAR      = "BAR"
	constValueFoo    = "foo"
)

var (
	testingDir     string
	err            error
	defaultBomFile = "../fakes/config/bom/tkg-bom-v1.3.0.yaml"
)

func TestTKGConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkg config updater Suite")
}

var _ = Describe("SaveConfig", func() {
	BeforeSuite((func() {
		testingDir = fakehelper.CreateTempTestingDirectory()
	}))

	AfterSuite((func() {
		fakehelper.DeleteTempTestingDirectory(testingDir)
	}))

	var (
		vars              map[string]string
		err               error
		clusterConfigPath string
		originalFile      []byte
		key               string
		value             string
	)
	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		originalFile, err = os.ReadFile(clusterConfigPath)
		Expect(err).ToNot(HaveOccurred())
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, vars)
	})

	Context("When the tkgconfig file contains the key", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyBAR
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
		})

		It("should override the key with the new value", func() {
			Expect(err).ToNot(HaveOccurred())
			res, err := getValue(clusterConfigPath, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(value))
		})
	})

	Context("When the tkgconfig file does not contains the key", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyFOO
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
		})

		It("should append the key-value pair to the tkgconfig vile", func() {
			Expect(err).ToNot(HaveOccurred())
			res, err := getValue(clusterConfigPath, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(value))
		})
	})

	Context("When the key is exported to environment variable", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyBAR
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
			_ = os.Setenv(key, "bar")
		})

		It("should override the environment variable with the new value", func() {
			Expect(err).ToNot(HaveOccurred())
			res := os.Getenv(key)
			Expect(res).To(Equal(value))
		})
	})

	Context("When the key is exported to environment variable", func() {
		BeforeEach(func() {
			clusterConfigPath = constConfigPath
			key = constKeyFOO
			value = constValueFoo
			vars = make(map[string]string)
			vars[key] = value
		})

		It("should export new key-value pair to environment variable", func() {
			Expect(err).ToNot(HaveOccurred())
			res := os.Getenv(key)
			Expect(res).To(Equal(value))
		})
	})

	AfterEach(func() {
		err = os.WriteFile(clusterConfigPath, originalFile, constants.ConfigFilePermissions)
		Expect(err).ToNot(HaveOccurred())

		_ = os.Unsetenv("FOO")
		_ = os.Unsetenv(constKeyBAR)
	})
})

var _ = Describe("Credential Encoding/Decoding", func() {
	var (
		clusterConfigPath     string
		client                Client
		tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		err                   error
	)

	BeforeEach(func() {
		createTempDirectory("reader_test")
	})

	JustBeforeEach(func() {
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		tkgConfigNode := loadTKGNode(clusterConfigPath)

		client.EnsureCredEncoding(tkgConfigNode)
		writeYaml(clusterConfigPath, tkgConfigNode)

		_, err = tkgconfigreaderwriter.New(clusterConfigPath)
		Expect(err).ToNot(HaveOccurred())
		err = client.DecodeCredentialsInViper()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("When the credential is in clear text format", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVspherePassword]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:QWRtaW4hMjM=>")) // base64encoded value of Admin!23

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableVspherePassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("Admin!23"))
		})
	})

	Context("When the credential is already base64 encoded", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableAWSAccessKeyID]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:UVdSRVRZVUlPUExLSkhHRkRTQVo=>")) // base64encoded value of QWRETYUIOPLKJHGFDSAZ

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableAWSAccessKeyID)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("QWRETYUIOPLKJHGFDSAZ"))
		})
	})

	Context("When the credential is in clear text and it has $ char in it", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableAWSSecretAccessKey]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:dU5uY0NhdEl2V3UxZSRycXdlcmtnMzVxVTdkc3dmRWE0cmRYSmsvRQ==>")) // base64encoded value of uNncCatIvWu1e$rqwerkg35qU7dswfEa4rdXJk/E

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableAWSSecretAccessKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("uNncCatIvWu1e$rqwerkg35qU7dswfEa4rdXJk/E"))
		})
	})

	Context("When the credential is in clear text format and passed through UI", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
			res := map[string]string{
				constants.ConfigVariableVspherePassword: "Admin$123",
			}
			tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())
			err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, res)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have encoded value in config file and clear text value in viper", func() {
			configByte, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configByte, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVspherePassword]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal("<encoded:QWRtaW4kMTIz>")) // base64encoded value of Admin$23

			viperValue, err := tkgConfigReaderWriter.Get(constants.ConfigVariableVspherePassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(viperValue).To(Equal("Admin$123"))
		})
	})

	Context("When the ssh key is longer than 80 chars", func() {
		var longSSHString string
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config4.yaml")
			longSSHString = "ssh 123456789012345678901234567890123456789012345678901234567890XXXXXX yy"
			res := map[string]string{
				constants.ConfigVariableVsphereSSHAuthorizedKey: longSSHString,
			}
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())
			err = SaveConfig(clusterConfigPath, tkgConfigReaderWriter, res)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have saved the long string in a single line", func() {
			configBytes, err := os.ReadFile(clusterConfigPath)
			Expect(err).ToNot(HaveOccurred())

			configMap := make(map[string]interface{})
			err = yaml.Unmarshal(configBytes, &configMap)
			Expect(err).ToNot(HaveOccurred())
			value, ok := configMap[constants.ConfigVariableVsphereSSHAuthorizedKey]
			Expect(ok).To(Equal(true))
			Expect(value).To(Equal(longSSHString))

			expectedLineValue := fmt.Sprintf("%s: %s", constants.ConfigVariableVsphereSSHAuthorizedKey, longSSHString)
			contains := bytes.Contains(configBytes, []byte(expectedLineValue))
			Expect(contains).To(Equal(true))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureTemplateFiles", func() {
	var (
		err        error
		needUpdate bool
		client     Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
	})

	JustBeforeEach(func() {
		err = client.EnsureTemplateFiles(needUpdate)
	})

	Context("When the providers folder does not exsit", func() {
		BeforeEach(func() {
			needUpdate = false
		})

		It("should explode the providers fold under $HOME/.tkg", func() {
			Expect(err).ToNot(HaveOccurred())
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName)
			_, err = os.Stat(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When the providers folder exsits, and there is no need for updating", func() {
		BeforeEach(func() {
			needUpdate = false
			err = os.MkdirAll(testingDir, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = client.SaveTemplateFiles(testingDir, false)
			Expect(err).ToNot(HaveOccurred())
			// make sure the providers is written properly
			needUpdate, err = client.CheckProvidersNeedUpdate()
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})

		It("should keeps the original providers fold under $HOME/.tkg", func() {
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName)
			_, err = os.Stat(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
			needUpdate, err = client.CheckProvidersNeedUpdate()
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})
	})

	Context("When the providers folder exists, and it needs to be updated", func() {
		BeforeEach(func() {
			needUpdate = true
			// write providers to tmp folder and modify providers.sha256sum
			err = os.MkdirAll(testingDir, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = client.SaveTemplateFiles(testingDir, false)
			Expect(err).ToNot(HaveOccurred())
			providerChecksumPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersChecksumFileName)
			err = os.WriteFile(providerChecksumPath, []byte("mismatched-checksum"), constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
			// make sure the providers is written properly
			needUpdate, err := client.CheckProvidersNeedUpdate()
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(true))
		})

		It("should update the providers folder under $HOME/.tkg", func() {
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName)
			_, err = os.Stat(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
			needUpdate, err := client.CheckProvidersNeedUpdate()
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("CheckProvidersNeedUpdate", func() {
	var (
		err        error
		needUpdate bool
		client     Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		configPath := constConfigPath
		setupPrerequsiteForTesting(configPath, testingDir, defaultBomFile)
		tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(configPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
	})
	JustBeforeEach(func() {
		needUpdate, err = client.CheckProvidersNeedUpdate()
	})

	Context("When providers directory does not exist", func() {
		It("should return false for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})
	})

	Context("When providers/providers.sha256sum is missing", func() {
		BeforeEach(func() {
			// write providers to tmp folder but remove the providers.sha256sum
			err = os.MkdirAll(testingDir, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = client.SaveTemplateFiles(testingDir, false)
			Expect(err).ToNot(HaveOccurred())
			providerConfigPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersChecksumFileName)
			err = os.Remove(providerConfigPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return true for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(true))
		})
	})

	Context("When the bundled and local providers.sha256sum are identical", func() {
		BeforeEach(func() {
			// write providers to tmp folder
			err = os.MkdirAll(testingDir, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = client.SaveTemplateFiles(testingDir, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return false for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(false))
		})
	})

	Context("When the bundled and local providers.sha256sum have difference", func() {
		BeforeEach(func() {
			// write providers to tmp folder but change the providers.sha256sum
			err = os.MkdirAll(testingDir, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = client.SaveTemplateFiles(testingDir, false)
			Expect(err).ToNot(HaveOccurred())
			providerChecksumPath := filepath.Join(testingDir, constants.LocalProvidersFolderName, constants.LocalProvidersChecksumFileName)
			err = os.WriteFile(providerChecksumPath, []byte("mismatched-checksum"), constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return true for needUpdate flag ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(needUpdate).To(Equal(true))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureProviders", func() {
	var (
		err               error
		needUpdate        bool
		clusterConfigPath string
		tkgConfigNode     *yaml.Node
		client            Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		client = New(testingDir, NewProviderTest(), nil)
		err = os.MkdirAll(testingDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())
		err = client.SaveTemplateFiles(testingDir, false)
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		tkgConfigNode = loadTKGNode(clusterConfigPath)
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		err = client.EnsureProviders(needUpdate, tkgConfigNode)
	})

	Context("When providers section is absent from the tkg config", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig2Path
		})

		It("should append the providers section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))
		})
	})

	Context("When there is no need for update provider, and provider section exists", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig3Path
		})
		It("should append the providers section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(2))
		})
	})

	Context("When the provider section needs to be updated", func() {
		BeforeEach(func() {
			needUpdate = true
			clusterConfigPath = constConfig3Path
		})
		It("should append the providers section to the tkg config file ", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
			Expect(index).ToNot(Equal(-1))

			numOfProviders, err := countProviders()
			Expect(err).ToNot(HaveOccurred())
			// numOfProviders(6) + 1 customized provider
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(numOfProviders + 1))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("EnsureImages", func() {
	var (
		err               error
		needUpdate        bool
		clusterConfigPath string
		tkgConfigNode     *yaml.Node
		client            Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
	})

	JustBeforeEach(func() {
		setupPrerequsiteForTesting(clusterConfigPath, testingDir, defaultBomFile)
		tkgConfigNode = loadTKGNode(clusterConfigPath)
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		err = client.EnsureImages(needUpdate, tkgConfigNode)
	})

	Context("when the images section is absent from the tkg config", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfigPath
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
		})
	})

	Context("when there is no need to update tkg config file", func() {
		BeforeEach(func() {
			needUpdate = false
			clusterConfigPath = constConfig2Path
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
			// 2 * (1 key node + 1 value node)
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(4))
		})
	})

	Context("when the images section needs to be updated", func() {
		BeforeEach(func() {
			needUpdate = true
			clusterConfigPath = constConfig2Path
		})

		It("should append the images section to the tkg config file", func() {
			Expect(err).ToNot(HaveOccurred())
			index := getNodeIndex(tkgConfigNode.Content[0].Content, constants.ImagesConfigKey)
			Expect(index).ToNot(Equal(-1))
			// 2 * (1 key node + 1 value node)
			Expect(tkgConfigNode.Content[0].Content[index].Content).To(HaveLen(4))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

var _ = Describe("CheckTkgConfigNeedUpdate", func() {
	var (
		err                 error
		clusterConfigPath   string
		tkgConfigNeedUpdate bool
		tkgConfigNode       *yaml.Node
		client              Client
	)

	BeforeEach(func() {
		createTempDirectory("template_test")
		client = New(testingDir, NewProviderTest(), nil)
		err = client.SaveTemplateFiles(testingDir, false)
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		out, werr := yaml.Marshal(tkgConfigNode)
		Expect(werr).ToNot(HaveOccurred())
		var tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
		tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigPath, filepath.Join(testingDir, "config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		client = New(testingDir, NewProviderTest(), tkgConfigReaderWriter)
		err = os.WriteFile(clusterConfigPath, out, 0o600)
		Expect(err).ToNot(HaveOccurred())
		tkgConfigNeedUpdate, _, err = client.CheckTkgConfigNeedUpdate()
	})

	Context("When the tkg config is up-to-date with the providers config", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config5.yaml")
			tkgConfigNode = loadTKGNode(clusterConfigPath)
			err = client.EnsureProviders(true, tkgConfigNode)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return false, indicating the tkgconfig is up-to-date", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(tkgConfigNeedUpdate).To(Equal(false))
		})
	})

	Context("When the providers section in tkg config is missing some entires", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config5.yaml")
			tkgConfigNode = loadTKGNode(clusterConfigPath)
		})
		It("should return true, indicating the tkgconfig need to be updated", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(tkgConfigNeedUpdate).To(Equal(true))
		})
	})

	Context("When some entries in the providers section in tkg config is different from the providers config", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config6.yaml")
			tkgConfigNode = loadTKGNode(clusterConfigPath)
		})

		It("should return true, indicating the tkgconfig needs to be updated", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(tkgConfigNeedUpdate).To(Equal(true))
		})
	})

	Context("When the providers sections is missing from the tkg config", func() {
		BeforeEach(func() {
			clusterConfigPath = getConfigFilePath("config_missing_providers.yaml")
			err = os.WriteFile(clusterConfigPath, []byte("bar: bar-val"), 0o600)
			Expect(err).ToNot(HaveOccurred())
			tkgConfigNode = loadTKGNode(clusterConfigPath)
		})

		It("should return false, indicating the tkgconfig doesn't need to be updated", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(tkgConfigNeedUpdate).To(Equal(false))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
	})
})

func getNodeIndex(node []*yaml.Node, key string) int {
	appIdx := -1
	for i, k := range node {
		if i%2 == 0 && k.Value == key {
			appIdx = i + 1
			break
		}
	}
	return appIdx
}

func getValue(filepath, key string) (string, error) {
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	tkgConfigNode := yaml.Node{}
	err = yaml.Unmarshal(fileData, &tkgConfigNode)
	if err != nil {
		return "", err
	}

	indexKey := getNodeIndex(tkgConfigNode.Content[0].Content, key)
	if indexKey == -1 {
		return "", errors.New("cannot find the key")
	}

	return tkgConfigNode.Content[0].Content[indexKey].Value, nil
}

func createTempDirectory(prefix string) {
	testingDir, err = os.MkdirTemp("", prefix)
	if err != nil {
		fmt.Println("Error TempDir: ", err.Error())
	}
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func getConfigFilePath(filename string) string {
	filePath := "../fakes/config/" + filename
	return setupPrerequsiteForTesting(filePath, testingDir, defaultBomFile)
}

func writeYaml(path string, tkgConfigNode *yaml.Node) {
	out, err := yaml.Marshal(tkgConfigNode)
	if err != nil {
		fmt.Println("Error marshaling tkg config to yaml", err.Error())
	}
	err = os.WriteFile(path, out, constants.ConfigFilePermissions)
	if err != nil {
		fmt.Println("Error WriteFile", err.Error())
	}
}

func loadTKGNode(path string) *yaml.Node {
	tkgConfigNode := yaml.Node{}
	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error ReadFile")
	}
	err = yaml.Unmarshal(fileData, &tkgConfigNode)
	if err != nil {
		fmt.Println("Error unmashaling tkg config")
	}

	return &tkgConfigNode
}

type provider struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	ProviderType string `yaml:"type"`
}

type providers struct {
	Providers []provider `yaml:"providers"`
}

func countProviders() (int, error) {
	path := filepath.Join(testingDir)

	providerConfigBytes, err := os.ReadFile(filepath.Join(path, constants.LocalProvidersFolderName, constants.LocalProvidersConfigFileName))
	if err != nil {
		return 0, err
	}

	providersConfig := providers{}
	err = yaml.Unmarshal(providerConfigBytes, &providersConfig)
	if err != nil {
		return 0, err
	}

	return len(providersConfig.Providers), nil
}

func setupPrerequsiteForTesting(clusterConfigFile string, testingDir string, defaultBomFile string) string {
	bomDir, err := tkgconfigpaths.New(testingDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	testClusterConfigFile := filepath.Join(testingDir, "config.yaml")
	os.Remove(testClusterConfigFile)
	log.Infof("utils.CopyFile( %s, %s)", clusterConfigFile, testClusterConfigFile)
	err = utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	tkgconfigpaths.TKGDefaultBOMImageTag = utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())
	return testClusterConfigFile
}

type providertest struct{}

// New returns provider client which implements provider interface
func NewProviderTest() providerinterface.ProviderInterface {
	return &providertest{}
}

func (p *providertest) GetProviderBundle() ([]byte, error) {
	return os.ReadFile("../fakes/providers/providers.zip")
}
