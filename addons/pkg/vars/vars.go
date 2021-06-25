// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vars contains variables used throughout the codebase.
package vars

var (
	// TKGAddonsNamespace is the TKG addon Namespace.
	TKGAddonsNamespace = "tkg-system"

	// TKGAddonsServiceAccount is the TKG addon ServiceAccount.
	TKGAddonsServiceAccount = "tkg-addons-app-sa"

	// TKGAddonsClusterRole is the TKG addon ClusterRole.
	TKGAddonsClusterRole = "tkg-addons-app-cluster-role"

	// TKGAddonsClusterRoleBinding is the TKG addon ClusterRoleBinding.
	TKGAddonsClusterRoleBinding = "tkg-addons-app-cluster-role-binding"

	// TKGAddonsImagePullPolicy is pull policy for TKG addon images.
	TKGAddonsImagePullPolicy = "IfNotPresent"

	// TKGCorePackageRepositoryName is the name of core package repository applied in the cluster
	TKGCorePackageRepositoryName = "core"
)
