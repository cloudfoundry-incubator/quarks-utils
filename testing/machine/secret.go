package machine

import (
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CreateSecret creates a secret and returns a function to delete it
func (m *Machine) CreateSecret(namespace string, secret corev1.Secret) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().Secrets(namespace)
	_, err := client.Create(&secret)
	return func() error {
		err := client.Delete(secret.GetName(), &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// UpdateSecret updates a secret and returns a function to delete it
func (m *Machine) UpdateSecret(namespace string, secret corev1.Secret) (*corev1.Secret, TearDownFunc, error) {
	client := m.Clientset.CoreV1().Secrets(namespace)
	s, err := client.Update(&secret)
	return s, func() error {
		err := client.Delete(secret.GetName(), &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// CollectSecret polls until the specified secret can be fetched
func (m *Machine) CollectSecret(namespace string, name string) (*corev1.Secret, error) {
	err := m.WaitForSecret(namespace, name)
	if err != nil {
		return nil, errors.Wrap(err, "waiting for secret "+name)
	}
	return m.GetSecret(namespace, name)
}

// GetSecret fetches the specified secret
func (m *Machine) GetSecret(namespace string, name string) (*corev1.Secret, error) {
	secret, err := m.Clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for secret "+name)
	}

	return secret, nil
}

// WaitForSecret blocks until the secret is available. It fails after the timeout.
func (m *Machine) WaitForSecret(namespace string, name string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.SecretExists(namespace, name)
	})
}

// WaitForSecretDeletion blocks until the resource is deleted
func (m *Machine) WaitForSecretDeletion(namespace string, name string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		found, err := m.SecretExists(namespace, name)
		return !found, err
	})
}

// SecretChangedFunc returns true if something changed in the secret
type SecretChangedFunc func(corev1.Secret) bool

// WaitForSecretChange blocks until the secret fulfills the change func
func (m *Machine) WaitForSecretChange(namespace string, name string, changed SecretChangedFunc) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		client := m.Clientset.CoreV1().Secrets(namespace)
		s, err := client.Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, errors.Wrapf(err, "failed to query for secret: %s", name)
		}

		return changed(*s), nil
	})
}

// SecretExists returns true if the secret by that name exist
func (m *Machine) SecretExists(namespace string, name string) (bool, error) {
	_, err := m.Clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for secret by name: %s", name)
	}

	return true, nil
}

// DeleteSecrets deletes all the secrets
func (m *Machine) DeleteSecrets(namespace string) (bool, error) {
	err := m.Clientset.CoreV1().Secrets(namespace).DeleteCollection(
		&metav1.DeleteOptions{},
		metav1.ListOptions{},
	)
	if err != nil {
		return false, errors.Wrapf(err, "failed to delete all secrets in namespace: %s", namespace)
	}

	return true, nil
}
