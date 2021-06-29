package lib

import (
	"bytes"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/wait"

	"time"

	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

type PackagePluginBase interface {
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	GetRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	ListRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult

	GetAvailablePackage(packageName string, o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult
	ListAvailablePackage(o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult

	CreateInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult
	GetInstalledPackage(o *tkgpackagedatamodel.PackageGetOptions) PackagePluginResult
	UpdateInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult
	DeleteInstalledPackage(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult
	ListInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult
}

type PackagePluginHelpers interface {
	AddOrUpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	CheckRepositoryAvailable(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	CheckAndDeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	CheckRepositoryDeleted(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult
	CheckPackageAvailable(packageName string, o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult
	CheckAndInstallPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult
	CheckPackageInstalled(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult
	CheckAndUninstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult
	CheckPackageDeleted(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult
}

type PackagePlugin interface {
	// Base commands are implemented by the below interface
	PackagePluginBase
	// Extra helper commands will be implemented by other interface
	PackagePluginHelpers
}

type PackagePluginResult struct {
	Pass   bool
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Error  error
}

type packagePlugin struct {
	kubeconfigPath string
	interval       time.Duration
	timeout        time.Duration
	outputFormat   string
	logFile        string
	verbose        int32
}

func NewPackagePlugin(kubeconfigPath string,
	interval time.Duration,
	timeout time.Duration,
	outputFormat string,
	logFile string,
	verbose int32) PackagePlugin {
	return &packagePlugin{
		kubeconfigPath: kubeconfigPath,
		interval:       interval,
		timeout:        timeout,
		outputFormat:   outputFormat,
		logFile:        logFile,
		verbose:        verbose,
	}
}

func (p *packagePlugin) addKubeconfig(cmd string) string {
	if cmd != "" && p.kubeconfigPath != "" {
		cmd += fmt.Sprintf(" --kubeconfig %s", p.kubeconfigPath)
	}
	return cmd
}

func (p *packagePlugin) addOutputFormat(cmd string) string {
	if cmd != "" && p.outputFormat != "" {
		cmd += fmt.Sprintf(" --output %s", p.outputFormat)
	}
	return cmd
}

func (p *packagePlugin) addLogFile(cmd string) string {
	if cmd != "" && p.logFile != "" {
		cmd += fmt.Sprintf(" --log-file %s", p.logFile)
	}
	return cmd
}

func (p *packagePlugin) addVerbose(cmd string) string {
	if cmd != "" && p.verbose != 0 {
		cmd += fmt.Sprintf(" --verbose %d", p.verbose)
	}
	return cmd
}

func (p *packagePlugin) addGlobalOptions(cmd string) string {
	cmd = p.addLogFile(cmd)
	cmd = p.addVerbose(cmd)
	return cmd
}

func (p *packagePlugin) AddRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository add %s %s", o.RepositoryName, o.RepositoryURL)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository get %s", o.RepositoryName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository update %s %s", o.RepositoryName, o.RepositoryURL)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateRepository {
		cmd += fmt.Sprintf(" --create")
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository delete %s", o.RepositoryName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.IsForceDelete {
		cmd += fmt.Sprintf(" --force")
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository list")
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetAvailablePackage(packageName string, o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package available get %s", packageName)

	if o.ValuesSchema {
		cmd += fmt.Sprintf(" --values-schema")
	}
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListAvailablePackage(o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package available list")
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) CreateInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed create %s --package-name %s", o.PkgInstallName, o.PackageName)
	if o.Version != "" {
		cmd += fmt.Sprintf(" --version %s", o.Version)
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	if o.ValuesFile != "" {
		cmd += fmt.Sprintf(" --values-file %s", o.ValuesFile)
	}
	if o.ServiceAccountName != "" {
		cmd += fmt.Sprintf(" --service-account-name %s", o.ServiceAccountName)
	}
	if !o.Wait {
		cmd += fmt.Sprintf(" --wait false")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %d", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %d", o.PollTimeout)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetInstalledPackage(o *tkgpackagedatamodel.PackageGetOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed get %s", o.PackageName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) UpdateInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed update %s --package-name %s", o.PkgInstallName, o.PackageName)
	if o.Version != "" {
		cmd += fmt.Sprintf(" --version %s", o.Version)
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	if o.ValuesFile != "" {
		cmd += fmt.Sprintf(" --values-file %s", o.ValuesFile)
	}
	if o.ServiceAccountName != "" {
		cmd += fmt.Sprintf(" --service-account-name %s", o.ServiceAccountName)
	}
	if !o.Wait {
		cmd += fmt.Sprintf(" --wait false")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %d", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %d", o.PollTimeout)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) DeleteInstalledPackage(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed delete %s", o.PkgInstallName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %d", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %d", o.PollTimeout)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListInstalledPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed list")
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" -A")
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) AddOrUpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetRepository(&tkgpackagedatamodel.RepositoryOptions{
		RepositoryName: o.RepositoryName,
		Namespace:      o.Namespace,
	})
	if getResult.Error != nil {
		return p.AddRepository(o)
	} else {
		if !strings.Contains(getResult.Stdout.String(), o.RepositoryURL) {
			return p.UpdateRepository(o)
		}
	}
	return result
}

func (p *packagePlugin) CheckRepositoryAvailable(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result := p.GetRepository(&tkgpackagedatamodel.RepositoryOptions{
			RepositoryName: o.RepositoryName,
			Namespace:      o.Namespace,
		})
		if result.Error != nil {
			return false, result.Error
		} else {
			if strings.Contains(result.Stdout.String(), "Reconcile succeeded") {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndDeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetRepository(&tkgpackagedatamodel.RepositoryOptions{
		RepositoryName: o.RepositoryName,
		Namespace:      o.Namespace,
	})
	if getResult.Error == nil {
		return p.DeleteRepository(o)
	}
	return result
}

func (p *packagePlugin) CheckRepositoryDeleted(o *tkgpackagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result := p.GetRepository(&tkgpackagedatamodel.RepositoryOptions{
			RepositoryName: o.RepositoryName,
			Namespace:      o.Namespace,
		})
		if result.Error != nil {
			// Setting result error to nil since there will be error on get after repository is deleted
			result.Error = nil
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckPackageAvailable(packageName string, o *tkgpackagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetAvailablePackage(packageName, o)
		if result.Error == nil {
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndInstallPackage(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetInstalledPackage(&tkgpackagedatamodel.PackageGetOptions{
		PackageName: o.PkgInstallName,
		Namespace:   o.Namespace,
	})
	if getResult.Error != nil {
		return p.CreateInstalledPackage(o)
	}
	return result
}

func (p *packagePlugin) CheckPackageInstalled(o *tkgpackagedatamodel.PackageInstalledOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetInstalledPackage(&tkgpackagedatamodel.PackageGetOptions{
			PackageName: o.PkgInstallName,
			Namespace:   o.Namespace,
		})
		if result.Error != nil {
			return false, result.Error
		} else {
			if strings.Contains(result.Stdout.String(), "Reconcile succeeded") {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndUninstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetInstalledPackage(&tkgpackagedatamodel.PackageGetOptions{
		PackageName: o.PkgInstallName,
		Namespace:   o.Namespace,
	})
	if getResult.Error == nil {
		return p.DeleteInstalledPackage(o)
	}
	return result
}

func (p *packagePlugin) CheckPackageDeleted(o *tkgpackagedatamodel.PackageUninstallOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetInstalledPackage(&tkgpackagedatamodel.PackageGetOptions{
			PackageName: o.PkgInstallName,
			Namespace:   o.Namespace,
		})
		if result.Error != nil {
			// Setting result error to nil since there will be error on get after package is deleted
			result.Error = nil
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}