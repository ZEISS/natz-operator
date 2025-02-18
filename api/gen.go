//go:build generate
// +build generate

//go:generate rm -rf ../manifests/crd/bases
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.3 object:headerFile="../hack/copyright.go.txt" paths="./..."
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.3 rbac:roleName=manager-role crd webhook output:crd:artifacts:config=../manifests/crd/bases paths="./..."

package api

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" //nolint:typecheck
)
