package versionedsecretstore

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
	"code.cloudfoundry.org/quarks-utils/pkg/pointers"
)

var (
	// LabelSecretKind is the label key for secret kind
	LabelSecretKind = fmt.Sprintf("%s/secret-kind", names.GroupName)
	// LabelVersion is the label key for secret version
	LabelVersion = fmt.Sprintf("%s/secret-version", names.GroupName)
	// LabelAPIVersion is the lable for kube APIVersion
	LabelAPIVersion = fmt.Sprintf("%s/v1alpha1", names.GroupName)
	// AnnotationSourceDescription is the label key for source description
	AnnotationSourceDescription = fmt.Sprintf("%s/source-description", names.GroupName)
)

const (
	// VersionSecretKind is the kind of versioned secret
	VersionSecretKind = "versionedSecret"
)

var _ VersionedSecretStore = &VersionedSecretImpl{}

// VersionedSecretStore is the interface to version secrets in Kubernetes
//
// Each update to the secret results in a new persisted version.
// An existing persisted version of a secret cannot be altered or deleted.
// The deletion of a secret will result in the removal of all persisted version of that secret.
//
// The version number is an integer that is incremented with each version of
// the secret, which the greatest number being the current/latest version.
//
// When saving a new secret, a source description is required, which
// should explain the sources of the rendered secret, e.g. the location of
// the Custom Resource Definition that generated it.
type VersionedSecretStore interface {
	SetSecretReferences(ctx context.Context, namespace string, podSpec *corev1.PodSpec) error
	Create(ctx context.Context, namespace string, ownerName string, ownerID types.UID, secretName string, secretData map[string]string, labels map[string]string, sourceDescription string) error
	Get(ctx context.Context, namespace string, secretName string, version int) (*corev1.Secret, error)
	Latest(ctx context.Context, namespace string, secretName string) (*corev1.Secret, error)
	List(ctx context.Context, namespace string, secretName string) ([]corev1.Secret, error)
	VersionCount(ctx context.Context, namespace string, secretName string) (int, error)
	Delete(ctx context.Context, namespace string, secretName string) error
	Decorate(ctx context.Context, namespace string, secretName string, key string, value string) error
}

// VersionedSecretImpl contains the required fields to persist a secret
type VersionedSecretImpl struct {
	client client.Client
}

// NewVersionedSecretStore returns a VersionedSecretStore implementation to be used
// when working with desired secret secrets
func NewVersionedSecretStore(client client.Client) VersionedSecretImpl {
	return VersionedSecretImpl{
		client,
	}
}

// SetSecretReferences update versioned secret references in pod spec
func (p VersionedSecretImpl) SetSecretReferences(ctx context.Context, namespace string, podSpec *corev1.PodSpec) error {
	_, secretsInSpec := GetConfigNamesFromSpec(*podSpec)
	for secretNameInSpec := range secretsInSpec {

		versionedSecretPrefix := NamePrefix(secretNameInSpec)
		// If this secret doesn't look like a versioned secret (e.g. <name>-v2), move on
		if versionedSecretPrefix == "" {
			continue
		}

		// We have the current secret name, we have to look and see if there's a new version
		versionedSecret, err := p.Latest(ctx, namespace, versionedSecretPrefix)

		// If the latest version of the secret doesn't exist yet, ignore this secret and move on
		// There should be no situation where a version n + 1 exists, and versions 0 through n don't exist
		if err != nil && apierrors.IsNotFound(err) {
			ctxlog.Debugf(ctx, "versioned secret %s in namespace %s doesn't exist", versionedSecretPrefix, namespace)
			continue
		}

		if err != nil {
			return errors.Wrapf(err, "failed to get latest versioned secret %s in namespace %s", versionedSecretPrefix, namespace)
		}

		// Make sure that the secret we're looking at is an actual versioned secret
		secretLabel := versionedSecret.Labels
		if secretLabel == nil {
			continue
		}

		secretKind, ok := secretLabel[LabelSecretKind]
		if !ok || secretKind != VersionSecretKind {
			continue
		}

		// if the latest version is different than the current version in the spec, replace it
		if versionedSecret.Name != secretNameInSpec {
			replaceVolumesSecretRef(
				podSpec.Volumes,
				secretNameInSpec,
				versionedSecret.GetName(),
			)

			replaceContainerEnvsSecretRef(
				podSpec.Containers,
				secretNameInSpec,
				versionedSecret.GetName(),
			)
		}
	}

	return nil
}

// Create creates a new version of the secret from secret data
func (p VersionedSecretImpl) Create(ctx context.Context, namespace string, ownerName string, ownerID types.UID, secretName string, secretData map[string]string, labels map[string]string, sourceDescription string) error {
	currentVersion, err := p.getGreatestVersion(ctx, namespace, secretName)
	if err != nil {
		return err
	}

	version := currentVersion + 1
	labels[LabelVersion] = strconv.Itoa(version)
	labels[LabelSecretKind] = VersionSecretKind

	generatedSecretName, err := GenerateSecretName(secretName, version)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generatedSecretName,
			Namespace: namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         LabelAPIVersion,
					Kind:               "ExtendedJob",
					Name:               ownerName,
					UID:                ownerID,
					BlockOwnerDeletion: pointers.Bool(false),
					Controller:         pointers.Bool(true),
				},
			},
			Annotations: map[string]string{
				AnnotationSourceDescription: sourceDescription,
			},
		},
		StringData: secretData,
	}

	return p.client.Create(ctx, secret)
}

// Get returns a specific version of the secret
func (p VersionedSecretImpl) Get(ctx context.Context, namespace string, deploymentName string, version int) (*corev1.Secret, error) {
	name, err := GenerateSecretName(deploymentName, version)
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	err = p.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// Latest returns the latest version of the secret
func (p VersionedSecretImpl) Latest(ctx context.Context, namespace string, secretName string) (*corev1.Secret, error) {
	latestVersion, err := p.getGreatestVersion(ctx, namespace, secretName)
	if err != nil {
		return nil, err
	}
	return p.Get(ctx, namespace, secretName, latestVersion)
}

// List returns all versions of the secret
func (p VersionedSecretImpl) List(ctx context.Context, namespace string, secretName string) ([]corev1.Secret, error) {
	secrets, err := p.listSecrets(ctx, namespace, secretName)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// VersionCount returns the number of versions for this secret
func (p VersionedSecretImpl) VersionCount(ctx context.Context, namespace string, secretName string) (int, error) {
	list, err := p.listSecrets(ctx, namespace, secretName)
	if err != nil {
		return 0, err
	}

	return len(list), nil
}

// Decorate adds a label to the latest version of the secret
func (p VersionedSecretImpl) Decorate(ctx context.Context, namespace string, secretName string, key string, value string) error {
	version, err := p.getGreatestVersion(ctx, namespace, secretName)
	if err != nil {
		return err
	}

	generatedSecretName, err := GenerateSecretName(secretName, version)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{}
	if err := p.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: generatedSecretName}, secret); err != nil {
		return err
	}

	labels := secret.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[key] = value
	secret.SetLabels(labels)

	return p.client.Update(ctx, secret)
}

// Delete removes all versions of the secret and therefore the
// secret itself.
func (p VersionedSecretImpl) Delete(ctx context.Context, namespace string, secretName string) error {
	list, err := p.listSecrets(ctx, namespace, secretName)
	if err != nil {
		return err
	}

	for _, secret := range list {
		if err := p.client.Delete(ctx, &secret); err != nil {
			return err
		}
	}

	return nil
}

func (p VersionedSecretImpl) listSecrets(ctx context.Context, namespace string, secretName string) ([]corev1.Secret, error) {
	secretLabelsSet := labels.Set{
		LabelSecretKind: VersionSecretKind,
	}

	secrets := &corev1.SecretList{}

	if err := p.client.List(
		ctx,
		secrets,
		client.InNamespace(namespace),
		client.MatchingLabels(secretLabelsSet),
	); err != nil {
		return nil, errors.Wrapf(err, "Failed to list secrets with labels %s", secretLabelsSet.String())
	}

	result := []corev1.Secret{}

	nameRegex := regexp.MustCompile(fmt.Sprintf(`^%s-v\d+$`, secretName))
	for _, secret := range secrets.Items {
		if nameRegex.MatchString(secret.Name) {
			result = append(result, secret)
		}
	}

	return result, nil
}

func (p VersionedSecretImpl) getGreatestVersion(ctx context.Context, namespace string, secretName string) (int, error) {
	list, err := p.listSecrets(ctx, namespace, secretName)
	if err != nil {
		return -1, err
	}

	var greatestVersion int
	for _, secret := range list {
		version, err := Version(secret)
		if err != nil {
			return 0, err
		}

		if version > greatestVersion {
			greatestVersion = version
		}
	}

	return greatestVersion, nil
}

// GenerateSecretName creates the name of a versioned secret and errors if it's invalid
func GenerateSecretName(namePrefix string, version int) (string, error) {
	proposedName := fmt.Sprintf("%s-v%d", namePrefix, version)

	// Check for Kubernetes name requirements (length)
	const maxChars = 253
	if len(proposedName) > maxChars {
		return "", errors.Errorf("secret name exceeds maximum number of allowed characters (actual=%d, allowed=%d)", len(proposedName), maxChars)
	}

	// Check for Kubernetes name requirements (characters)
	if re := regexp.MustCompile(`[^a-z0-9.-]`); re.MatchString(proposedName) {
		return "", errors.Errorf("secret name contains invalid characters, only lower case, dot and dash are allowed")
	}

	return proposedName, nil
}

// replaceVolumesSecretRef replace secret reference of volumes
func replaceVolumesSecretRef(volumes []corev1.Volume, secretName string, versionedSecretName string) {
	for _, vol := range volumes {
		if vol.VolumeSource.Secret != nil && vol.VolumeSource.Secret.SecretName == secretName {
			vol.VolumeSource.Secret.SecretName = versionedSecretName
		}
	}
}

// replaceContainerEnvsSecretRef replace secret reference of envs for each container
func replaceContainerEnvsSecretRef(containers []corev1.Container, secretName string, versionedSecretName string) {
	for _, container := range containers {

		for _, env := range container.EnvFrom {
			if s := env.SecretRef; s != nil {
				if s.Name == secretName {
					s.Name = versionedSecretName
				}
			}
		}

		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}
			if sRef := env.ValueFrom.SecretKeyRef; sRef != nil {
				if sRef.Name == secretName {
					sRef.Name = versionedSecretName
				}
			}
		}
	}
}
