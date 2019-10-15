package versionedsecretstore

import (
       "golang.org/x/net/context"

       corev1 "k8s.io/api/core/v1"
       "k8s.io/apimachinery/pkg/types"
       "sigs.k8s.io/controller-runtime/pkg/client"
)

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