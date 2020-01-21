package machine

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"code.cloudfoundry.org/quarks-utils/pkg/pointers"
)

// Machine produces and destroys resources for tests
type Machine struct {
	PollTimeout  time.Duration
	PollInterval time.Duration

	Clientset *kubernetes.Clientset
}

// TearDownFunc tears down the resource
type TearDownFunc func() error

// ChanResult holds different fields that can be
// sent through a channel
type ChanResult struct {
	Error error
}

const (
	// DefaultTimeout used to wait for resources
	DefaultTimeout = 300 * time.Second
	// DefaultInterval for polling
	DefaultInterval = 500 * time.Millisecond
)

// NewMachine returns a new machine which creates k8s resources
func NewMachine() Machine {
	return Machine{
		PollTimeout:  DefaultTimeout,
		PollInterval: DefaultInterval,
	}
}

// CreateNamespace creates a namespace, it doesn't return an error if the namespace exists
func (m *Machine) CreateNamespace(namespace string) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().Namespaces()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.Create(ns)
	if apierrors.IsAlreadyExists(err) {
		err = nil
	}
	return func() error {
		b := metav1.DeletePropagationBackground
		err := client.Delete(ns.GetName(), &metav1.DeleteOptions{
			// this is run in aftersuite before failhandler, so let's keep the namespace for a few seconds
			GracePeriodSeconds: pointers.Int64(5),
			PropagationPolicy:  &b,
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// ContainExpectedEvent return true if events contain target resource event
func (m *Machine) ContainExpectedEvent(events []corev1.Event, reason string, message string) bool {
	for _, event := range events {
		if event.Reason == reason && strings.Contains(event.Message, message) {
			return true
		}
	}

	return false
}

// TearDownAll calls all passed in tear down functions in order
func (m *Machine) TearDownAll(funcs []TearDownFunc) error {
	var messages string
	for _, f := range funcs {
		err := f()
		if err != nil {
			messages = fmt.Sprintf("%v%v\n", messages, err.Error())
		}
	}
	if messages != "" {
		return errors.New(messages)
	}
	return nil
}

// GetService gets target Service
func (m *Machine) GetService(namespace string, name string) (*corev1.Service, error) {
	svc, err := m.Clientset.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return svc, errors.Wrapf(err, "failed to get service '%s'", svc)
	}

	return svc, nil
}

// CreateService creates a Service in the given namespace
func (m *Machine) CreateService(namespace string, service corev1.Service) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().Services(namespace)
	_, err := client.Create(&service)
	return func() error {
		err := client.Delete(service.GetName(), &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// WaitForPortReachable blocks until the endpoint is reachable
func (m *Machine) WaitForPortReachable(protocol, uri string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		_, err := net.Dial(protocol, uri)
		return err == nil, nil
	})
}

// GetEndpoints gets target Endpoints
func (m *Machine) GetEndpoints(namespace string, name string) (*corev1.Endpoints, error) {
	ep, err := m.Clientset.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return ep, errors.Wrapf(err, "failed to get endpoint '%s'", ep)
	}

	return ep, nil
}

// WaitForSubsetsExist blocks until the specified endpoints' subsets exist
func (m *Machine) WaitForSubsetsExist(namespace string, endpointsName string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		found, err := m.SubsetsExist(namespace, endpointsName)
		return found, err
	})
}

// SubsetsExist checks if the subsets of the endpoints exist
func (m *Machine) SubsetsExist(namespace string, endpointsName string) (bool, error) {
	ep, err := m.Clientset.CoreV1().Endpoints(namespace).Get(endpointsName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for endpoints by endpointsName: %s", endpointsName)
	}

	if len(ep.Subsets) == 0 {
		return false, nil
	}

	return true, nil
}

// GetNodes gets nodes
func (m *Machine) GetNodes() ([]corev1.Node, error) {
	nodes := []corev1.Node{}

	nodeList, err := m.Clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nodes, nil
		}
		return nodes, errors.Wrapf(err, "failed to query for nodes")
	}

	if len(nodeList.Items) == 0 {
		return nodes, nil
	}

	nodes = nodeList.Items

	return nodes, nil
}
