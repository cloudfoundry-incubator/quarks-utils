// Package crd handles the creation and updating of our CRDs in the cluster
package crd

import (
	"context"
	"reflect"
	"time"

	"code.cloudfoundry.org/quarks-utils/pkg/pointers"
	"github.com/pkg/errors"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Builder builds CRDs
type Builder struct {
	crdName                  string
	names                    extv1.CustomResourceDefinitionNames
	groupVersion             schema.GroupVersion
	validation               *extv1.CustomResourceValidation
	additionalPrinterColumns []extv1.CustomResourceColumnDefinition
	CRD                      *extv1.CustomResourceDefinition
}

// New returns a new CRD builder
func New(
	crdName string,
	names extv1.CustomResourceDefinitionNames,
	groupVersion schema.GroupVersion,
) *Builder {
	return &Builder{
		crdName:      crdName,
		names:        names,
		groupVersion: groupVersion,
	}
}

// WithValidation add validation struct to the CRDs field
func (b *Builder) WithValidation(validation *extv1.CustomResourceValidation) *Builder {
	b.validation = validation
	return b
}

// WithAdditionalPrinterColumns add additional printer columns to the kubectl output
func (b *Builder) WithAdditionalPrinterColumns(cols []extv1.CustomResourceColumnDefinition) *Builder {
	b.additionalPrinterColumns = cols
	return b
}

// Build the CRD
func (b *Builder) Build() *Builder {
	b.CRD = &extv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.crdName,
		},
		Spec: extv1.CustomResourceDefinitionSpec{
			Group:                 b.groupVersion.Group,
			Names:                 b.names,
			Scope:                 extv1.NamespaceScoped,
			PreserveUnknownFields: pointers.Bool(false),
			Versions: []extv1.CustomResourceDefinitionVersion{
				{
					Name:    b.groupVersion.Version,
					Served:  true,
					Storage: true,
				},
			},
			Validation: b.validation,
			Subresources: &extv1.CustomResourceSubresources{
				Status: &extv1.CustomResourceSubresourceStatus{},
			},
			AdditionalPrinterColumns: b.additionalPrinterColumns,
		},
	}
	return b
}

// Apply CRD to cluster
func (b *Builder) Apply(ctx context.Context, client extv1client.ApiextensionsV1beta1Interface) error {
	existing, err := client.CustomResourceDefinitions().Get(ctx, b.crdName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "getting CRD '%s'", b.crdName)
		}
		_, err := client.CustomResourceDefinitions().Create(ctx, b.CRD, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "creating CRD '%s'", b.crdName)
		}
		return nil
	}

	if !reflect.DeepEqual(b.CRD.Spec, existing.Spec) {
		b.CRD.ResourceVersion = existing.ResourceVersion
		_, err = client.CustomResourceDefinitions().Update(ctx, b.CRD, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "updating CRD '%s'", b.crdName)
		}
	}

	return nil
}

// ApplyCRD creates or updates the CRD - old func for compatibility
func ApplyCRD(ctx context.Context, client extv1client.ApiextensionsV1beta1Interface, crdName, kind, plural string, shortNames []string, groupVersion schema.GroupVersion, validation *extv1.CustomResourceValidation) error {
	b := New(
		crdName,
		extv1.CustomResourceDefinitionNames{
			Kind:       kind,
			Plural:     plural,
			ShortNames: shortNames,
		},
		groupVersion,
	)
	return b.WithValidation(validation).Build().Apply(ctx, client)
}

// WaitForCRDReady blocks until the CRD is ready.
func WaitForCRDReady(ctx context.Context, client extv1client.ApiextensionsV1beta1Interface, crdName string) error {
	err := wait.ExponentialBackoff(
		wait.Backoff{
			Duration: time.Second,
			Steps:    15,
			Factor:   1,
		},
		func() (bool, error) {
			crd, err := client.CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			for _, cond := range crd.Status.Conditions {
				if cond.Type == extv1.NamesAccepted && cond.Status == extv1.ConditionTrue {
					return true, nil
				}
			}

			return false, nil
		})
	if err != nil {
		return errors.Wrapf(err, "Waiting for CRD ready failed")
	}

	return nil
}
