package versionedsecretstore

import (
	"golang.org/x/net/context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type versionedSecretStoreClientsetBackend struct {
	clientset kubernetes.Interface
}

func (b *versionedSecretStoreClientsetBackend) Create(ctx context.Context, secret *corev1.Secret) error {
	_, err := b.clientset.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (b *versionedSecretStoreClientsetBackend) Get(ctx context.Context, nn types.NamespacedName) (*corev1.Secret, error) {
	return b.clientset.CoreV1().Secrets(nn.Namespace).Get(ctx, nn.Name, metav1.GetOptions{})
}

func (b *versionedSecretStoreClientsetBackend) Update(ctx context.Context, secret *corev1.Secret) error {
	_, err := b.clientset.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (b *versionedSecretStoreClientsetBackend) Delete(ctx context.Context, secret *corev1.Secret) error {
	err := b.clientset.CoreV1().Secrets(secret.Namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
	return err
}

func (b *versionedSecretStoreClientsetBackend) List(ctx context.Context, namespace string, matchLabels map[string]string) (*corev1.SecretList, error) {

	secrets, err := b.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(matchLabels).String(),
	})
	return secrets, err
}

type versionedSecretStoreClientBackend struct {
	client client.Client
}

func (b *versionedSecretStoreClientBackend) Create(ctx context.Context, secret *corev1.Secret) error {
	return b.client.Create(ctx, secret)
}

func (b *versionedSecretStoreClientBackend) Get(ctx context.Context, nn types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := b.client.Get(ctx, nn, secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (b *versionedSecretStoreClientBackend) Update(ctx context.Context, secret *corev1.Secret) error {
	return b.client.Update(ctx, secret)
}

func (b *versionedSecretStoreClientBackend) Delete(ctx context.Context, secret *corev1.Secret) error {
	return b.client.Delete(ctx, secret)
}

func (b *versionedSecretStoreClientBackend) List(ctx context.Context, namespace string, matchLabels map[string]string) (*corev1.SecretList, error) {
	secrets := &corev1.SecretList{}

	err := b.client.List(
		ctx,
		secrets,
		client.InNamespace(namespace),
		client.MatchingLabels(matchLabels),
	)
	return secrets, err
}
