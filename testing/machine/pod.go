package machine

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	podutil "code.cloudfoundry.org/quarks-utils/pkg/pod"
)

// CreatePod creates a default pod and returns a function to delete it
func (m *Machine) CreatePod(namespace string, pod corev1.Pod) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().Pods(namespace)
	_, err := client.Create(context.Background(), &pod, metav1.CreateOptions{})
	return func() error {
		err := client.Delete(context.Background(), pod.GetName(), metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// WaitForPod blocks until the pod is running. It fails after the timeout.
func (m *Machine) WaitForPod(namespace string, name string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.PodRunning(namespace, name)
	})
}

// WaitForPodReady blocks until the pod is ready. It fails after the timeout.
func (m *Machine) WaitForPodReady(namespace string, name string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.PodReady(namespace, name)
	})
}

// WaitForPods blocks until all selected pods are running. It fails after the timeout.
func (m *Machine) WaitForPods(namespace string, labels string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.PodsRunning(namespace, labels)
	})
}

// WaitForPodFailures blocks until all selected pods are failing. It fails after the timeout.
func (m *Machine) WaitForPodFailures(namespace string, labels string) error {
	return wait.PollImmediate(5*time.Second, m.PollTimeout, func() (bool, error) {
		return m.PodsFailing(namespace, labels)
	})
}

// WaitForInitContainerRunning blocks until a pod's init container is running.
// It fails after the timeout.
func (m *Machine) WaitForInitContainerRunning(namespace, podName, containerName string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.InitContainerRunning(namespace, podName, containerName)
	})
}

// WaitForPodsDelete blocks until the pod is deleted. It fails after the timeout.
func (m *Machine) WaitForPodsDelete(namespace string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		return m.PodsDeleted(namespace)
	})
}

// PodsDeleted returns true if the all pods are deleted
func (m *Machine) PodsDeleted(namespace string) (bool, error) {
	podList, err := m.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if len(podList.Items) == 0 {
		return true, nil
	}
	return false, nil
}

// PodRunning returns true if the pod by that name is in state running
func (m *Machine) PodRunning(namespace string, name string) (bool, error) {
	pod, err := m.Clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for pod by name: %s", name)
	}

	if pod.Status.Phase == corev1.PodRunning {
		return true, nil
	}
	return false, nil
}

// PodReady returns true if the pod by that name is ready.
func (m *Machine) PodReady(namespace string, name string) (bool, error) {
	pod, err := m.Clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for pod by name: %s", name)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return false, nil
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}

// InitContainerRunning returns true if the pod by that name has a specific init container that is in state running
func (m *Machine) InitContainerRunning(namespace, podName, containerName string) (bool, error) {
	pod, err := m.Clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for pod by name: %s", podName)
	}

	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.Name != containerName {
			continue
		}

		if containerStatus.State.Running != nil {
			return true, nil
		}
	}

	return false, nil
}

// PodsFailing returns true if the pod by that name exist and is in a failed state
func (m *Machine) PodsFailing(namespace string, labels string) (bool, error) {
	pods, err := m.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to query for pod by labels: %v", labels)
	}

	if len(pods.Items) == 0 {
		return false, nil
	}

	for _, pod := range pods.Items {

		pos, condition := podutil.GetPodCondition(&pod.Status, corev1.ContainersReady)
		if (pos > -1 && condition.Reason == "ContainersNotReady") ||
			pod.Status.Phase == corev1.PodFailed {

			return true, nil
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			state := containerStatus.State
			if (state.Waiting != nil && state.Waiting.Reason == "ImagePullBackOff") ||
				(state.Waiting != nil && state.Waiting.Reason == "ErrImagePull") {
				return true, nil
			}
		}
	}

	return false, nil
}

// PodsRunning returns true if all the pods selected by labels are in state running
// Note that only the first page of pods is considered - don't use this if you have a
// long pod list that you care about
func (m *Machine) PodsRunning(namespace string, labels string) (bool, error) {
	pods, err := m.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to query for pod by labels: %v", labels)
	}

	if len(pods.Items) == 0 {
		return false, nil
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
	}

	return true, nil
}

// PodCount returns the number of matching pods
func (m *Machine) PodCount(namespace string, labels string, match func(corev1.Pod) bool) (int, error) {
	pods, err := m.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to query for pod by labels: %v", labels)
	}

	for _, pod := range pods.Items {
		if !match(pod) {
			return -1, nil
		}
	}

	return len(pods.Items), nil
}

// GetPods returns all the pods selected by labels
func (m *Machine) GetPods(namespace string, labels string) (*corev1.PodList, error) {
	pods, err := m.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		return &corev1.PodList{}, errors.Wrapf(err, "failed to query for pod by labels: %v", labels)
	}

	return pods, nil

}

// GetPod returns pod by name
func (m *Machine) GetPod(namespace string, name string) (*corev1.Pod, error) {
	pod, err := m.Clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query for pod by name: %v", name)
	}

	return pod, nil
}

// UpdatePod updates a pod and returns a function to delete it
func (m *Machine) UpdatePod(namespace string, pod corev1.Pod) (*corev1.Pod, TearDownFunc, error) {
	client := m.Clientset.CoreV1().Pods(namespace)
	s, err := client.Update(context.Background(), &pod, metav1.UpdateOptions{})
	return s, func() error {
		err := client.Delete(context.Background(), pod.GetName(), metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}, err
}

// PodLabeled returns true if the pod is labeled correctly
func (m *Machine) PodLabeled(namespace string, name string, desiredLabel, desiredValue string) (bool, error) {
	pod, err := m.Clientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, err
		}
		return false, errors.Wrapf(err, "failed to query for pod by name: %s", name)
	}

	if pod.ObjectMeta.Labels[desiredLabel] == desiredValue {
		return true, nil
	}
	return false, fmt.Errorf("cannot match the desired label with %s", desiredValue)
}
