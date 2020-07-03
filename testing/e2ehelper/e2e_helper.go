package e2ehelper

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/pkg/errors"

	"code.cloudfoundry.org/quarks-utils/testing"
)

const (
	installTimeOutInSecs = "600"
	e2eFailedMessage     = "Failed setting up e2e test environment."
)

var (
	nsIndex int
)

// TearDownFunc tears down the resource
type TearDownFunc func() error

// TearDownAll calls all passed in tear down functions in order
func TearDownAll(funcs []TearDownFunc) error {
	var messages string
	for _, f := range funcs {
		if f != nil {
			err := f()
			if err != nil {
				messages = fmt.Sprintf("%v%v\n", messages, err.Error())
			}
		}
	}
	if messages != "" {
		return errors.New(messages)
	}
	return nil
}

// CreateNamespace creates the operator namespace and returns the generated single namespace name
func CreateNamespace() (string, string, TearDownFunc, error) {
	operatorNamespace, err := createTestNamespace()
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "%s Creating test namespace failed.", e2eFailedMessage)
	}

	namespace := fmt.Sprintf("%s-work", operatorNamespace)

	fmt.Printf("Using test namespace: %s, %s\n", operatorNamespace, namespace)

	teardownFunc := func() error {
		return testing.DeleteNamespace(operatorNamespace)
	}

	return namespace, operatorNamespace, teardownFunc, nil
}

// createTestNamespace generates a namespace based on a env variable
func createTestNamespace() (string, error) {
	prefix, found := os.LookupEnv("TEST_NAMESPACE")
	if !found {
		prefix = "default"
	}
	namespace := prefix + "-" + strconv.Itoa(config.GinkgoConfig.ParallelNode) + "-" + strconv.Itoa(int(nsIndex))
	nsIndex++

	err := testing.CreateNamespace(namespace)
	if err != nil {
		return "", err
	}

	return namespace, nil
}

// CreateMonitoredNamespaceFromExistingRole creates a namespace with the monitored label
func CreateMonitoredNamespaceFromExistingRole(clusterRole string) (string, []TearDownFunc, error) {
	prefix, found := os.LookupEnv("TEST_NAMESPACE")
	if !found {
		prefix = "default"
	}
	newNamespace := prefix + "-multins-" + strconv.Itoa(config.GinkgoConfig.ParallelNode) + "-" + strconv.Itoa(int(nsIndex))
	roleName := prefix + "-multins-role-" + strconv.Itoa(config.GinkgoConfig.ParallelNode) + "-" + strconv.Itoa(int(nsIndex))
	nsIndex++

	err := testing.CreateNamespace(newNamespace)
	if err != nil {
		return "", nil, err
	}

	kubectl := testing.NewKubectl()
	err = kubectl.CreateServiceAccount(newNamespace, newNamespace)
	if err != nil {
		return "", nil, err
	}

	err = kubectl.CreateRoleBinding(newNamespace, clusterRole, newNamespace+":"+newNamespace, roleName)
	if err != nil {
		return "", nil, err
	}

	err = testing.PatchNamespace(newNamespace,
		`[{"op": "add", "path": "/metadata/labels", "value": {"quarks.cloudfoundry.org/qjob-service-account": "`+newNamespace+`", "quarks.cloudfoundry.org/monitored": "`+clusterRole+`"}}]`)

	if err != nil {
		return "", nil, err
	}

	f := []TearDownFunc{
		func() error { return kubectl.DeleteRoleBinding(newNamespace, roleName) },
		func() error { return kubectl.DeleteServiceAccount(newNamespace, newNamespace) },
		func() error { return testing.DeleteNamespace(newNamespace) },
	}

	return newNamespace, f, nil
}

// CreateMonitoredNamespace creates a namespace with the monitored label
func CreateMonitoredNamespace(namespace string, id string) (TearDownFunc, error) {
	err := testing.CreateNamespace(namespace)
	if err != nil {
		return nil, err
	}

	err = testing.PatchNamespace(namespace,
		`[{"op": "add", "path": "/metadata/labels", "value": {"quarks.cloudfoundry.org/monitored": "`+id+`"}}]`)
	if err != nil {
		return nil, err
	}

	f := func() error {
		return testing.DeleteNamespace(namespace)
	}
	return f, nil
}

// InstallChart installs the helm chart into the operator namespace
func InstallChart(chartPath string, operatorNamespace string, args ...string) (TearDownFunc, error) {
	err := testing.RunHelmBinaryWithCustomErr("version")
	if err != nil {
		return nil, errors.Wrapf(err, "%s Helm version command failed", e2eFailedMessage)
	}

	cmd := append([]string{
		"install",
		fmt.Sprintf("%s-%s", testing.CFOperatorRelease, operatorNamespace),
		chartPath,
		"--namespace", operatorNamespace,
		"--timeout", fmt.Sprintf("%ss", installTimeOutInSecs),
		"--wait",
	}, args...)

	err = testing.RunHelmBinaryWithCustomErr(cmd...)
	if err != nil {
		return nil, errors.Wrapf(err, "%s Helm install command failed.", e2eFailedMessage)
	}

	// Add sleep for workaround for CI timeouts
	time.Sleep(10 * time.Second)

	return func() error {
		err = testing.RunHelmBinaryWithCustomErr("delete", "-n", operatorNamespace, fmt.Sprintf("%s-%s", testing.CFOperatorRelease, operatorNamespace))
		if err != nil {
			return err
		}

		return nil
	}, nil
}
