// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package azure

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for Azure - Antrea", func() {
	E2ECommonSpec(context.TODO(), func() E2ECommonSpecInput {
		return E2ECommonSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
		}
	})
})
