package config

import (
	"fmt"

	"code.cloudfoundry.org/quarks-utils/pkg/names"
	corev1 "k8s.io/api/core/v1"
)

var config struct {
	// operatorDockerImage is the location of the operators own docker image
	operatorDockerImage     string
	operatorImagePullPolicy corev1.PullPolicy
}

func init() {
	config.operatorImagePullPolicy = corev1.PullIfNotPresent
}

// SetupOperatorDockerImage initializes the package scoped variable
func SetupOperatorDockerImage(org, repo, tag string) (err error) {
	config.operatorDockerImage, err = names.GetDockerSourceName(org, repo, tag)
	return
}

// GetOperatorDockerImage returns the image name of the operator docker image
func GetOperatorDockerImage() string {
	return config.operatorDockerImage
}

// SetupOperatorImagePullPolicy sets the pull policy
func SetupOperatorImagePullPolicy(pullPolicy string) error {
	switch pullPolicy {
	case string(corev1.PullAlways):
		config.operatorImagePullPolicy = corev1.PullAlways
	case string(corev1.PullNever):
		config.operatorImagePullPolicy = corev1.PullNever
	case string(corev1.PullIfNotPresent):
		config.operatorImagePullPolicy = corev1.PullIfNotPresent
	default:
		return fmt.Errorf("invalid image pull policy '%s'", pullPolicy)
	}
	return nil
}

// GetOperatorImagePullPolicy returns the image pull policy to be used for generated pods
func GetOperatorImagePullPolicy() corev1.PullPolicy {
	return config.operatorImagePullPolicy
}
