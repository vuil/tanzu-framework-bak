package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonconfig "github.com/vmware-tanzu-private/core/addons/pkg/config"
	addonconstants "github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	"github.com/vmware-tanzu-private/core/addons/pkg/util"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type AppReconciler struct {
	Config addonconfig.Config
}

// nolint:funlen
func (r AppReconciler) ReconcileAddonKappResourceNormal(
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

		/*
		 * remoteApp means App CR on the management cluster that kapp-controller uses to remotely manages set of objects deployed in a workload cluster.
		 * workload clusters kubeconfig details need to be added for remote App so that kapp-controller on management
		 * cluster can reconcile and push the addon/app to the workload cluster
		 */
		if remoteApp {
			clusterKubeconfigDetails := util.GetClusterKubeconfigSecretDetails(remoteCluster)

			app.Spec.Cluster = &kappctrl.AppCluster{
				KubeconfigSecretRef: &kappctrl.AppClusterKubeconfigSecretRef{
					Name: clusterKubeconfigDetails.Name,
					Key:  clusterKubeconfigDetails.Key,
				},
			}
		} else {
			app.Spec.ServiceAccountName = addonconstants.TKGAddonsAppServiceAccount
		}

		app.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.AppSyncPeriod}

		templateImageURL, err := util.GetTemplateImageURLFromBom(addonConfig, imageRepository, bom)
		if err != nil {
			log.Error(err, "Error getting addon template image")
			return err
		}
		log.Info("Addon template image found", constants.ImageURLLogKey, templateImageURL)

		app.Spec.Fetch = []kappctrl.AppFetch{
			{
				Image: &kappctrl.AppFetchImage{
					URL: templateImageURL,
				},
			},
		}

		app.Spec.Template = []kappctrl.AppTemplate{
			{
				Ytt: &kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
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
		if remoteApp {
			app.Spec.NoopDelete = true
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, app, appMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon App")
		return err
	}

	logOperationResult(log, "app", result)

	return nil
}

// nolint:dupl
func (r AppReconciler) ReconcileAddonKappResourceDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	if err := clusterClient.Delete(ctx, app); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Addon app not found")
			return nil
		}
		log.Error(err, "Error deleting addon app")
		return err
	}

	log.Info("Deleted app")

	return nil
}

func (r AppReconciler) ReconcileAddonDataValuesSecretNormal(
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
			addonDataValuesSecret.Data[k] = v
		}
		// Add or updates the imageInfo if container image reference exists
		if len(addonConfig.AddonContainerImages) > 0 {
			imageInfoBytes, err := util.GetImageInfo(addonConfig, imageRepository, bom)
			if err != nil {
				log.Error(err, "Error retrieving addon image info")
				return err
			}
			addonDataValuesSecret.Data["imageInfo.yaml"] = imageInfoBytes
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	logOperationResult(log, "addon app data values secret", result)

	return nil
}
