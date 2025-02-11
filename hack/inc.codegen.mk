# Code generation
#
# see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#generate-code

# Name of the Go package for this repository
PKG 		:= github.com/zeiss/natz-operator

# List of API groups to generate code for
# e.g. "v1alpha1 v1alpha2"
API_GROUPS 	:=  v1alpha1
# generates e.g. "PKG/api/v1alpha1 PKG/api/v1alpha2"
api-import-paths := $(foreach group,$(API_GROUPS),$(PKG)/api/$(group))

generators 	:= deepcopy client

.PHONY: codegen $(generators)
codegen: $(generators)

# http://blog.jgc.org/2007/06/escaping-comma-and-space-in-gnu-make.html
comma := ,
null  :=
space := $(null) $(null)

client:
	@echo "+ Generating clientsets for $(API_GROUPS)"
	@rm -rf pkg/client/generated/clientset
		echo "+ Generating clientsets for $$apigrp" ; \
		$(GO_RUN_TOOLS) k8s.io/code-generator/cmd/client-gen \
			--fake-clientset=true \
			--input $(subst $(space),$(comma),$(API_GROUPS)) \
			--input-base $(PKG)/api \
			--go-header-file hack/copyright.go.txt \
			--output-pkg $(PKG)/pkg/client/generated/clientset \
			--output-dir pkg/client/generated/clientset; \

# Cleanup codegen
.PHONY: codegen-cleanup
codegen-cleanup:
	@if [ -d "./$(PKG)" ]; then \
		cp -a ./$(PKG)/pkg/client/generated/ pkg/client/generated/ ;\
		cp -a ./$(PKG)/apis/* apis/ ;\
		rm -rf "./$(PKG)" ;\
	fi