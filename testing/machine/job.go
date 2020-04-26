package machine

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// DeleteJobs deletes all the jobs
func (m *Machine) DeleteJobs(namespace string, labels string) (bool, error) {
	err := m.Clientset.BatchV1().Jobs(namespace).DeleteCollection(
		context.Background(),
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: labels},
	)
	if err != nil {
		return false, errors.Wrapf(err, "failed to delete all jobs with labels: %s", labels)
	}

	return true, nil
}

// WaitForJobsDeleted waits until the jobs no longer exists
func (m *Machine) WaitForJobsDeleted(namespace string, labels string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		jobs, err := m.Clientset.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
		})
		if err != nil {
			return false, errors.Wrapf(err, "failed to list jobs by label: %s", labels)
		}

		return len(jobs.Items) < 1, nil
	})
}

// JobExists returns true if job with that name exists
func (m *Machine) JobExists(namespace string, name string) (bool, error) {
	_, err := m.Clientset.BatchV1().Jobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to query for job by name: %s", name)
	}

	return true, nil
}

// CollectJobs waits for n jobs with specified labels.
// It fails after the timeout.
func (m *Machine) CollectJobs(namespace string, labels string, n int) ([]batchv1.Job, error) {
	found := map[string]batchv1.Job{}
	err := wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		jobs, err := m.Clientset.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
		})
		if err != nil {
			return false, errors.Wrapf(err, "failed to query for jobs by label: %s", labels)
		}

		for _, job := range jobs.Items {
			found[job.GetName()] = job
		}
		return len(found) >= n, nil
	})

	if err != nil {
		return nil, err
	}

	jobs := []batchv1.Job{}
	for _, job := range found {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// WaitForJobExists polls until a short timeout is reached or a job is found
// It returns true only if a job is found
func (m *Machine) WaitForJobExists(namespace string, labels string) (bool, error) {
	found := false
	err := wait.Poll(5*time.Second, 30*time.Second, func() (bool, error) {
		jobs, err := m.Clientset.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
		})
		if err != nil {
			return false, errors.Wrapf(err, "failed to query for jobs by label: %s", labels)
		}

		found = len(jobs.Items) != 0
		return found, err
	})

	if err != nil && strings.Contains(err.Error(), "timed out waiting for the condition") {
		err = nil
	}

	return found, err
}

// WaitForJobDeletion blocks until the batchv1.Job is deleted
func (m *Machine) WaitForJobDeletion(namespace string, name string) error {
	return wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		found, err := m.JobExists(namespace, name)
		return !found, err
	})
}

// ContainJob searches job array for a job matching `name`
func (m *Machine) ContainJob(jobs []batchv1.Job, name string) bool {
	for _, job := range jobs {
		if strings.Contains(job.GetName(), name) {
			return true
		}
	}
	return false
}
