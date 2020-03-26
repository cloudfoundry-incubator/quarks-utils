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

const (
	// ConfigMap is used in log messages
	ConfigMap = "configmap"
	// Secret is used in log messages
	Secret = "secret"
)

var allowedDNSLabelChars = regexp.MustCompile("[^-a-z0-9]*")

// DNSLabelSafe returns a string which is safe to use as a DNS label
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
func DNSLabelSafe(name string) string {
	name = strings.Replace(name, "_", "-", -1)
	name = strings.ToLower(name)
	name = allowedDNSLabelChars.ReplaceAllLiteralString(name, "")
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, "-")
	return name
}

// Sanitize produces valid k8s names, i.e. for containers: [a-z0-9]([-a-z0-9]*[a-z0-9])?
// The result is DNS label compatible and usable as a path segment name.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
func Sanitize(name string) string {
	name = DNSLabelSafe(name)
	return TruncateMD5(name, 63)
}

var allowedDNSSubdomainChars = regexp.MustCompile("[^-.a-z0-9]*")

// SanitizeSubdomain allows more than Sanitize, cannot be used for DNS or path segments.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
func SanitizeSubdomain(name string) string {
	name = strings.Replace(name, "_", "-", -1)
	name = strings.ToLower(name)
	name = allowedDNSSubdomainChars.ReplaceAllLiteralString(name, "")
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, "-")
	name = strings.TrimPrefix(name, ".")
	name = strings.TrimSuffix(name, ".")
	name = TruncateMD5(name, 253)
	return name
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

// VolumeName generate volume name based on secret name
func VolumeName(secretName string) string {
	nameSlices := strings.Split(secretName, ".")
	volName := ""
	if len(nameSlices) > 1 {
		volName = nameSlices[1]
	} else {
		volName = nameSlices[0]
	}
	return Sanitize(volName)
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

// TruncateMD5 truncates the string after n chars and add a hex encoded md5
// sum, to produce a uniq representation of the original string.  Example:
// names are limited to 63 characters so we recalculate the name as <name
// trimmed to 31 characters>-<md5 hash of name>
func TruncateMD5(s string, n int) string {
	if len(s) > n {
		sumHex := md5.Sum([]byte(s))
		sum := hex.EncodeToString(sumHex[:])
		s = s[:n-32] + sum
	}
	return s
}
