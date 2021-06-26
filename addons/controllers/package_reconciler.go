package controllers

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu-private/core/addons/pkg/vars"

	addonconfig "github.com/vmware-tanzu-private/core/addons/pkg/config"
	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type PackageReconciler struct {
	Config addonconfig.Config
}

// reconcileCorePackageRepository reconciles the core package repository in the cluster
func (r PackageReconciler) reconcileCorePackageRepository(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	imageRepository string,
	bom *bomtypes.Bom) error {

	repositoryImage, err := util.GetCorePackageRepositoryImageFromBom(bom)
	if err != nil {
		log.Error(err, "Core package repository image not found", constants.PackageRepositoryLogKey, constants.TKGCorePackageRepositoryImageName)
		return err
	}

	// build the core PackageRepository CR
	corePackageRepository := &pkgiv1alpha1.PackageRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vars.TKGCorePackageRepositoryName,
			Namespace: vars.TKGAddonsNamespace,
		},
	}

	// apply the core PackageRepository CR
	addonDataValuesSecretMutateFn := func() error {
		corePackageRepository.Spec = pkgiv1alpha1.PackageRepositorySpec{
			Fetch: &pkgiv1alpha1.PackageRepositoryFetch{
				ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{
					Image: fmt.Sprintf("%s/%s:%s", imageRepository, repositoryImage.ImagePath, repositoryImage.Tag),
				},
			},
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, corePackageRepository, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching core package repository")
		return err
	}
	logOperationResult(log, "core package repository", result)

	return nil
}

func (r PackageReconciler) ReconcileAddonKappResourceNormal(
	ctx context.Context,
	log logr.Logger,
	remoteApp bool,
	remoteCluster *clusterapiv1alpha3.Cluster,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	/*
	 * remoteApp means App CR on the management cluster that kapp-controller uses to remotely manages set of objects deployed in a workload cluster.
	 * workload clusters kubeconfig details need to be added for remote App so that kapp-controller on management
	 * cluster can reconcile and push the addon/app to the workload cluster
	 */
	if remoteApp {
		// TODO: Switch to remote PackageInstall when this feature is available in packaging api

		app := &kappctrl.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
				Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
			},
		}

		appMutateFn := func() error {
			if app.ObjectMeta.Annotations == nil {
				app.ObjectMeta.Annotations = make(map[string]string)
			}

			app.ObjectMeta.Annotations[addontypes.AddonTypeAnnotation] = fmt.Sprintf("%s/%s", addonConfig.Category, addonName)
			app.ObjectMeta.Annotations[addontypes.AddonNameAnnotation] = addonSecret.Name
			app.ObjectMeta.Annotations[addontypes.AddonNamespaceAnnotation] = addonSecret.Namespace

			clusterKubeconfigDetails := util.GetClusterKubeconfigSecretDetails(remoteCluster)

			app.Spec.Cluster = &kappctrl.AppCluster{
				KubeconfigSecretRef: &kappctrl.AppClusterKubeconfigSecretRef{
					Name: clusterKubeconfigDetails.Name,
					Key:  clusterKubeconfigDetails.Key,
				},
			}

			app.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.AppSyncPeriod}

			templateImageURL, err := util.GetTemplateImageURLFromBom(addonConfig, imageRepository, bom)
			if err != nil {
				log.Error(err, "Error getting addon template image")
				return err
			}
			log.Info("Addon template image found", constants.ImageURLLogKey, templateImageURL)

			// Use ImgpkgBundle in App CR
			app.Spec.Fetch = []kappctrl.AppFetch{
				{
					ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{
						Image: templateImageURL,
					},
				},
			}

			app.Spec.Template = []kappctrl.AppTemplate{
				{
					Ytt: &kappctrl.AppTemplateYtt{
						IgnoreUnknownComments: true,
						Strict:                false,
						Paths:                 []string{"config"},
						Inline: &kappctrl.AppFetchInline{
							PathsFrom: []kappctrl.AppFetchInlineSource{
								{
									SecretRef: &kappctrl.AppFetchInlineSourceRef{
										Name: util.GenerateAppSecretNameFromAddonSecret(addonSecret),
									},
								},
							},
						},
					},
					Kbld: &kappctrl.AppTemplateKbld{
						Paths: []string{
							"-",
							".imgpkg/images.yml",
						},
					},
				},
			}

			app.Spec.Deploy = []kappctrl.AppDeploy{
				{
					Kapp: &kappctrl.AppDeployKapp{
						// --wait-timeout flag specifies the maximum time to wait for App deployment. In some corner cases,
						// current App could have the dependency on the deployment of another App, so current App could get
						// stuck in wait phase.
						RawOptions: []string{fmt.Sprintf("--wait-timeout=%s", r.Config.AppWaitTimeout)},
					},
				},
			}
			// If its a remoteApp set delete to no-op since the app doesnt have to be deleted when cluster is deleted.
			app.Spec.NoopDelete = true

			return nil
		}

		result, err := controllerutil.CreateOrPatch(ctx, clusterClient, app, appMutateFn)
		if err != nil {
			log.Error(err, "Error creating or patching addon remote App")
			return err
		}

		logOperationResult(log, "app", result)
	} else {
		ipkg := &pkgiv1alpha1.PackageInstall{
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
				Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
			},
		}

		ipkgMutateFn := func() error {
			if ipkg.ObjectMeta.Annotations == nil {
				ipkg.ObjectMeta.Annotations = make(map[string]string)
			}

			ipkg.ObjectMeta.Annotations[addontypes.AddonTypeAnnotation] = fmt.Sprintf("%s/%s", addonConfig.Category, addonName)
			ipkg.ObjectMeta.Annotations[addontypes.AddonNameAnnotation] = addonSecret.Name
			ipkg.ObjectMeta.Annotations[addontypes.AddonNamespaceAnnotation] = addonSecret.Namespace

			ipkg.Spec = pkgiv1alpha1.PackageInstallSpec{
				ServiceAccountName: vars.TKGAddonsServiceAccount,
				PackageRef: &pkgiv1alpha1.PackageRef{
					RefName: addonConfig.PackageName,
					VersionSelection: &versions.VersionSelectionSemver{
						Prereleases: &versions.VersionSelectionSemverPrereleases{},
					},
				},
				Values: []pkgiv1alpha1.PackageInstallValues{
					{SecretRef: &pkgiv1alpha1.PackageInstallValuesSecretRef{Name: util.GenerateAppSecretNameFromAddonSecret(addonSecret)}},
				},
			}

			return nil
		}

		result, err := controllerutil.CreateOrPatch(ctx, clusterClient, ipkg, ipkgMutateFn)
		if err != nil {
			log.Error(err, "Error creating or patching addon PackageInstall")
			return err
		}

		logOperationResult(log, "PackageInstall", result)
	}

	return nil
}

// nolint:dupl
func (r PackageReconciler) ReconcileAddonKappResourceDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	pkgi := &pkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	if err := clusterClient.Delete(ctx, pkgi); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Addon PackageInstall not found")
			return nil
		}
		log.Error(err, "Error deleting addon PackageInstall")
		return err
	}

	log.Info("Deleted PackageInstall")

	return nil
}

func (r PackageReconciler) ReconcileAddonDataValuesSecretNormal(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppSecretNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	addonDataValuesSecretMutateFn := func() error {
		addonDataValuesSecret.Type = corev1.SecretTypeOpaque
		addonDataValuesSecret.Data = map[string][]byte{}
		for k, v := range addonSecret.Data {
			// Trim the annotations if we are using the packaging API
			addonDataValuesSecret.Data[k] = util.TrimAddonDataValueAnnotations(v)
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	logOperationResult(log, "addon package data values secret", result)

	return nil
}
