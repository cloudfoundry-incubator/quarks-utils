package names

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DeploymentSecretType lists all the types of secrets used in
// the lifecycle of a BOSHDeployment
type DeploymentSecretType int

const (
	// ConfigMap is used in log messages
	ConfigMap = "configmap"
	// Secret is used in log messages
	Secret = "secret"
)

const (
	// DeploymentSecretTypeManifestWithOps is a manifest that has ops files applied
	DeploymentSecretTypeManifestWithOps DeploymentSecretType = iota
	// DeploymentSecretTypeDesiredManifest is a manifest whose variables have been interpolated
	DeploymentSecretTypeDesiredManifest
	// DeploymentSecretTypeVariable is a BOSH variable generated using an QuarksSecret
	DeploymentSecretTypeVariable
	// DeploymentSecretTypeInstanceGroupResolvedProperties is a YAML file containing all properties needed to render an Instance Group
	DeploymentSecretTypeInstanceGroupResolvedProperties
	// DeploymentSecretBpmInformation is a YAML file containing the BPM information for one instance group
	DeploymentSecretBpmInformation
)

func (s DeploymentSecretType) String() string {
	return [...]string{
		"with-ops",
		"desired",
		"var",
		"ig-resolved",
		"bpm"}[s]
}

// DesiredManifestPrefix returns the prefix of the desired manifest's name:
func DesiredManifestPrefix(deploymentName string) string {
	return Sanitize(deploymentName) + "."
}

// DesiredManifestName returns the versioned name of the desired manifest
// secret, e.g. 'test.desired-manifest-v1'
func DesiredManifestName(deploymentName string, version string) string {
	finalName := DesiredManifestPrefix(deploymentName) + "desired-manifest"
	if version != "" {
		finalName = fmt.Sprintf("%s-v%s", finalName, version)
	}

	return finalName
}

// EntanglementSecretName returns the name of a secret containing properties of
// the provides section from a BOSH job
func EntanglementSecretName(deploymentName, igName string) string {
	return Sanitize(fmt.Sprintf("link-%s-%s", deploymentName, igName))
}

// EntanglementSecretKey returns the key (composed of type and name) for the k8s secret's data
func EntanglementSecretKey(linkType, linkName string) string {
	return fmt.Sprintf("%s.%s", linkType, linkName)
}

var secretNameRegex = regexp.MustCompile("[^-][a-z0-9-]*.[a-z0-9-]*[^-]")
var secretPartRegex = regexp.MustCompile("[a-z0-9-]*")

// DeploymentSecretName generates a Secret name for a given name and a deployment
func DeploymentSecretName(secretType DeploymentSecretType, deploymentName, name string) string {
	if name == "" {
		name = secretType.String()
	} else {
		name = fmt.Sprintf("%s-%s", secretType, name)
	}

	deploymentName = secretPartRegex.FindString(strings.Replace(deploymentName, "_", "-", -1))
	variableName := secretPartRegex.FindString(strings.Replace(name, "_", "-", -1))
	secretName := secretNameRegex.FindString(deploymentName + "." + variableName)

	return truncateMD5(secretName)
}

// DeploymentSecretPrefix returns the prefix used for our k8s secrets:
// `<deployment-name>.<secretType>.
func DeploymentSecretPrefix(secretType DeploymentSecretType, deploymentName string) string {
	return DeploymentSecretName(secretType, deploymentName, "") + "."
}

// InstanceGroupSecretName returns the name of a k8s secret:
// `<deployment-name>.<secretType>.<instance-group>-v<version>` secret.
//
// These secrets are created by QuarksJob and mounted on containers, e.g.
// for the template rendering.
func InstanceGroupSecretName(secretType DeploymentSecretType, deploymentName string, igName string, version string) string {
	prefix := DeploymentSecretPrefix(secretType, deploymentName)
	finalName := prefix + Sanitize(igName)

	if version != "" {
		finalName = fmt.Sprintf("%s-v%s", finalName, version)
	}

	return finalName
}

var allowedKubeChars = regexp.MustCompile("[^-a-z0-9]*")

// Sanitize produces valid k8s names, i.e. for containers: [a-z0-9]([-a-z0-9]*[a-z0-9])?
func Sanitize(name string) string {
	name = strings.Replace(name, "_", "-", -1)
	name = strings.ToLower(name)
	name = allowedKubeChars.ReplaceAllLiteralString(name, "")
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, "-")
	name = truncateMD5(name)
	return name
}

func truncateMD5(s string) string {
	if len(s) > 63 {
		// names are limited to 63 characters so we recalculate the name as
		// <name trimmed to 31 characters>-<md5 hash of name>
		sumHex := md5.Sum([]byte(s))
		sum := hex.EncodeToString(sumHex[:])
		s = s[:63-32] + sum
	}
	return s
}

// GetStatefulSetName gets statefulset name from podName
func GetStatefulSetName(name string) string {
	nameSplit := strings.Split(name, "-")
	nameSplit = nameSplit[0 : len(nameSplit)-1]
	statefulSetName := strings.Join(nameSplit, "-")
	return statefulSetName
}

// JobName returns a unique, short name for a given eJob k8s allows 63 chars,
// but the job's pod will have -\d{6} (=7 chars) appended.  So we return max 56
// chars: name39-suffix16
func JobName(eJobName string) (string, error) {
	name := truncate(eJobName, 39)

	hashID, err := randSuffix(name)
	if err != nil {
		return "", errors.Wrapf(err, "could not randomize job suffix for name %s", name)
	}
	return fmt.Sprintf("%s-%s", name, hashID), nil
}

var podOrdinalRegex = regexp.MustCompile(`(.*)-([0-9]+)$`)

// OrdinalFromPodName returns ordinal from pod name
func OrdinalFromPodName(name string) int {
	podOrdinal := -1
	match := podOrdinalRegex.FindStringSubmatch(name)
	if len(match) < 3 {
		return podOrdinal
	}
	if i, err := strconv.ParseInt(match[2], 10, 32); err == nil {
		podOrdinal = int(i)
	}
	return podOrdinal
}

// CSRName returns a CertificateSigningRequest name for a given QuarksJob
func CSRName(namespace, quarksSecretName string) string {
	return fmt.Sprintf("%s-%s", truncate(namespace, 19), Sanitize(truncate(quarksSecretName, 19)))
}

// CsrPrivateKeySecretName returns a Secret name for a given CertificateSigningRequest private key
func CsrPrivateKeySecretName(csrName string) string {
	return csrName + "-csr-private-key"
}

func randSuffix(str string) (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", errors.Wrap(err, "could not read rand bytes")
	}

	a := fnv.New64()
	_, err = a.Write([]byte(str + string(randBytes)))
	if err != nil {
		return "", errors.Wrapf(err, "could not write hash for str %s", str)
	}

	return hex.EncodeToString(a.Sum(nil)), nil
}

func truncate(name string, max int) string {
	if len(name) > max {
		return name[0:max]
	}
	return name
}
