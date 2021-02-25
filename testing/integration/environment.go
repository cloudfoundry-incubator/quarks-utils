// Package environment adds everything around mgr.Start() to run a local operator for the integration test suites
package environment

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" //from https://github.com/kubernetes/client-go/issues/345
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/testing/machine"
)

// Environment starts our operator and handles interaction with the k8s
// cluster used in the tests
type Environment struct {
	ID           int
	KubeConfig   *rest.Config
	Log          *zap.SugaredLogger
	TeardownFunc machine.TearDownFunc
	Config       *config.Config
	ObservedLogs *observer.ObservedLogs
	Namespace    string
	ctx          context.Context
	cancel       context.CancelFunc
}

// StartManager is used to clean up the test environment
func (e *Environment) StartManager(mgr manager.Manager) {
	e.ctx, e.cancel = context.WithCancel(context.Background())
	go func() {
		defer ginkgo.GinkgoRecover()
		gomega.Expect(mgr.Start(e.ctx)).NotTo(gomega.HaveOccurred())
	}()
}

// Teardown is used to clean up the test environment
func (e *Environment) Teardown(wasFailure bool) {
	if wasFailure {
		DumpENV(e.Namespace)
	}

	e.cancel()

	gexec.Kill()

	if e.TeardownFunc == nil {
		return
	}
	err := e.TeardownFunc()
	if err != nil && !NamespaceDeletionInProgress(err) {
		fmt.Printf("WARNING: failed to delete namespace %s: %v\n", e.Namespace, err)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

// KubeConfig returns a kube config for this environment
func KubeConfig() (*rest.Config, error) {
	location := os.Getenv("KUBECONFIG")
	if location == "" {
		location = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", location)
	if err != nil {
		log.Printf("INFO: cannot use kube config: %s\n", err)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
