//go:build generate
// +build generate

//go:generate rm -rf ../manifests/crd/bases
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.3 object:headerFile="../hack/copyright.go.txt" paths="./..."
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.3 rbac:roleName=manager-role crd webhook output:crd:artifacts:config=../manifests/crd/bases paths="./..."
//go:generate cp ../manifests/crd/bases/natz.zeiss.com_natsaccounts.yaml ../helm/charts/natz-operator/templates/crds/natsaccounts.yaml
//go:generate cp ../manifests/crd/bases/natz.zeiss.com_natsclusters.yaml ../helm/charts/natz-operator/templates/crds/natsoperators.yaml
//go:generate cp ../manifests/crd/bases/natz.zeiss.com_natsstreamingclusters.yaml ../helm/charts/natz-operator/templates/crds/natsusers.yaml

package api

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" //nolint:typecheck
)
