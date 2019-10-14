package cmd

import (
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

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

	return nil
}

// DockerImageFlags adds to viper flags
func DockerImageFlags(pf *flag.FlagSet, argToEnv map[string]string) {
	pf.StringP("docker-image-org", "o", "cfcontainerization", "Dockerhub organization that provides the operator docker image")
	pf.StringP("docker-image-repository", "r", "cf-operator", "Dockerhub repository that provides the operator docker image")
	pf.StringP("docker-image-tag", "t", "", "Tag of the operator docker image")

	viper.BindPFlag("docker-image-org", pf.Lookup("docker-image-org"))
	viper.BindPFlag("docker-image-repository", pf.Lookup("docker-image-repository"))
	viper.BindPFlag("docker-image-tag", pf.Lookup("docker-image-tag"))

	argToEnv["docker-image-org"] = "DOCKER_IMAGE_ORG"
	argToEnv["docker-image-repository"] = "DOCKER_IMAGE_REPOSITORY"
	argToEnv["docker-image-tag"] = "DOCKER_IMAGE_TAG"
}
