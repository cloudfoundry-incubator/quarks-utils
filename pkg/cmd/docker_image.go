package cmd

import (
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
)

// DockerImage sets the docker image, which the operator uses to execute commands in pods
func DockerImage() error {
	dockerImageTag := viper.GetString("docker-image-tag")
	if dockerImageTag == "" {
		return errors.Errorf("environment variable DOCKER_IMAGE_TAG not set")
	}

	err := config.SetupOperatorDockerImage(
		viper.GetString("docker-image-org"),
		viper.GetString("docker-image-repository"),
		dockerImageTag,
	)
	if err != nil {
		return errors.Wrap(err, "Couldn't parse docker image reference.")
	}

	err = config.SetupOperatorImagePullPolicy(viper.GetString("docker-image-pull-policy"))
	if err != nil {
		return errors.Wrap(err, "Couldn't parse docker image pull policy.")
	}

	return nil
}

// DockerImageFlags adds to viper flags
func DockerImageFlags(pf *flag.FlagSet, argToEnv map[string]string, repo string, tag string) {
	pf.StringP("docker-image-org", "o", "cfcontainerization", "Dockerhub organization that provides the operator docker image")
	pf.StringP("docker-image-repository", "r", repo, "Dockerhub repository that provides the operator docker image")
	pf.StringP("docker-image-tag", "t", tag, "Tag of the operator docker image")
	pf.StringP("docker-image-pull-policy", "", string(corev1.PullIfNotPresent), "Image pull policy")

	// TODO Check all of viper.BindPFlag on project level
	viper.BindPFlag("docker-image-org", pf.Lookup("docker-image-org"))                 //nolint:errcheck
	viper.BindPFlag("docker-image-repository", pf.Lookup("docker-image-repository"))   //nolint:errcheck
	viper.BindPFlag("docker-image-tag", pf.Lookup("docker-image-tag"))                 //nolint:errcheck
	viper.BindPFlag("docker-image-pull-policy", pf.Lookup("docker-image-pull-policy")) //nolint:errcheck

	argToEnv["docker-image-org"] = "DOCKER_IMAGE_ORG"
	argToEnv["docker-image-repository"] = "DOCKER_IMAGE_REPOSITORY"
	argToEnv["docker-image-tag"] = "DOCKER_IMAGE_TAG"
	argToEnv["docker-image-pull-policy"] = "DOCKER_IMAGE_PULL_POLICY"
}
