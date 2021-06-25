// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/kappclient"
)

var packageInstalledListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Args:  cobra.NoArgs,
	RunE:  packageInstalledList,
}

func init() {
	packageInstalledListCmd.Flags().BoolVarP(&packageinstalledOp.AllNamespaces, "all-namespaces", "A", false, "If present, list packages across all namespaces.")
	packageInstalledCmd.AddCommand(packageInstalledListCmd)
}

func packageInstalledList(cmd *cobra.Command, args []string) error {
	kc, err := kappclient.NewKappClient(packageinstalledOp.KubeConfig)
	if err != nil {
		return err
	}
	if packageinstalledOp.AllNamespaces {
		packageinstalledOp.Namespace = ""
	}
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving installed packages...", true)
	if err != nil {
		return err
	}

	pkgInstalledList, err := kc.ListPackageInstalls(packageinstalledOp.Namespace)
	if err != nil {
		return err
	}

	if packageinstalledOp.AllNamespaces {
		t.SetKeys("NAME", "PACKAGE-NAME", "PACKAGE-VERSION", "STATUS", "NAMESPACE")
	} else {
		t.SetKeys("NAME", "PACKAGE-NAME", "PACKAGE-VERSION", "STATUS")
	}
	for i := range pkgInstalledList.Items {
		pkg := pkgInstalledList.Items[i]
		if packageinstalledOp.AllNamespaces {
			t.AddRow(pkg.Name, pkg.Spec.PackageRef.RefName, pkg.Spec.PackageRef.VersionSelection.Constraints,
				pkg.Status.FriendlyDescription, pkg.Namespace)
		} else {
			t.AddRow(pkg.Name, pkg.Spec.PackageRef.RefName, pkg.Spec.PackageRef.VersionSelection.Constraints,
				pkg.Status.FriendlyDescription)
		}
	}
	t.RenderWithSpinner()
	return nil
}