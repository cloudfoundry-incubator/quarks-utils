package e2ehelper

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/pkg/errors"

	"code.cloudfoundry.org/quarks-utils/testing"
)

const (
	installTimeOutInSecs = "600"
	e2eFailedMessage     = "e2e test setting up environment failed."
)

var (
	nsIndex   int
	namespace string
)

// TearDownFunc tears down the resource
type TearDownFunc func() error

// SetUpEnvironment ensures helm binary can run,
// being able to reach tiller, and eventually it
// will install the cf-operator chart.
func SetUpEnvironment(chartPath string) (string, string, TearDownFunc, error) {
	// Ensure tiller is there, if not
	// then create it, via "init"
	err := testing.RunHelmBinaryWithCustomErr("version", "-s")
	if err != nil {
		switch err := err.(type) {
		case *testing.CustomError:
			if strings.Contains(err.StdOut, "could not find tiller") {
				err := testing.RunHelmBinaryWithCustomErr("init", "--wait")
				if err != nil {
					return "", "", nil, errors.Wrapf(err, "%s Helm init --wait command failed.", e2eFailedMessage)
				}
			}
		default:
			return "", "", nil, errors.Wrapf(err, "%s Helm version -s command failed", e2eFailedMessage)
		}
	}

	operatorNamespace, err := CreateTestNamespace()
	namespace = fmt.Sprintf("%s-work", operatorNamespace)
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "%s Creating test namespace failed.", e2eFailedMessage)
	}
	fmt.Println("Setting up in test namespace '" + namespace + "'...")

	// TODO: find relative path here
	err = testing.RunHelmBinaryWithCustomErr("install", chartPath,
		"--name", fmt.Sprintf("%s-%s", testing.CFOperatorRelease, operatorNamespace),
		"--namespace", operatorNamespace,
		"--timeout", installTimeOutInSecs,
		"--set", fmt.Sprintf("global.operator.watchNamespace=%s", namespace),
		"--wait")
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "%s Helm install command failed.", e2eFailedMessage)
	}

	// Add sleep for workaround for CI timeouts
	time.Sleep(10 * time.Second)

	teardownFunc := func() error {
		var messages string
		err = testing.DeleteSecret(namespace, "cf-operator-webhook-server-cert")
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}

		err = testing.DeleteWebhooks(namespace)
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}

		err := testing.RunHelmBinaryWithCustomErr("delete", fmt.Sprintf("%s-%s", testing.CFOperatorRelease, operatorNamespace), "--purge")
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}

		err = testing.DeleteNamespace(namespace)
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}

		err = testing.DeleteNamespace(operatorNamespace)
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}

		if messages != "" {
			fmt.Printf("Failures while cleaning up test environment for '%s':\n %v", namespace, messages)
			return errors.New(messages)
		}
		fmt.Println("Cleaned up test environment for '" + namespace + "'")
		return nil
	}

	return namespace, operatorNamespace, teardownFunc, nil
}

// CreateTestNamespace generates a namespace based on a env variable
func CreateTestNamespace() (string, error) {
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
