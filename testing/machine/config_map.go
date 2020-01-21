package machine

import (
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CreateConfigMap creates a ConfigMap and returns a function to delete it
func (m *Machine) CreateConfigMap(namespace string, configMap corev1.ConfigMap) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().ConfigMaps(namespace)
	_, err := client.Create(&configMap)
	return func() error {
		err := client.Delete(configMap.GetName(), &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// UpdateConfigMap updates a ConfigMap and returns a function to delete it
func (m *Machine) UpdateConfigMap(namespace string, configMap corev1.ConfigMap) (*corev1.ConfigMap, TearDownFunc, error) {
	client := m.Clientset.CoreV1().ConfigMaps(namespace)
	cm, err := client.Update(&configMap)
	return cm, func() error {
		err := client.Delete(configMap.GetName(), &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// GetConfigMap gets a ConfigMap by name
func (m *Machine) GetConfigMap(namespace string, name string) (*corev1.ConfigMap, error) {
	configMap, err := m.Clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return &corev1.ConfigMap{}, errors.Wrapf(err, "failed to query for configMap by name: %v", name)
	}

	return configMap, nil
}

// WaitForConfigMap blocks until the config map is available. It fails after the timeout.
func (m *Machine) WaitForConfigMap(namespace string, name string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.ConfigMapExists(namespace, name)
	})
}

// ConfigMapExists returns true if the secret by that name exist
func (m *Machine) ConfigMapExists(namespace string, name string) (bool, error) {
	_, err := m.Clientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for config map by name: %s", name)
	}

	return true, nil
}
