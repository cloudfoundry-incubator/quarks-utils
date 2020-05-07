// +build tools

// Dummy package to pin kubernetes/code-generator
// from go.mod used by 'make kube-gen' in quarks-operator
// and quarks-jobs.
// Required to allow "go mod vendor" to fetch the same kubecodegen
// version pinned in the go.mod in this repository

package codegen

import 	_ "k8s.io/code-generator" 
