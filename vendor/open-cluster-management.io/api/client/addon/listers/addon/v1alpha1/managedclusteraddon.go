// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
)

// ManagedClusterAddOnLister helps list ManagedClusterAddOns.
// All objects returned here must be treated as read-only.
type ManagedClusterAddOnLister interface {
	// List lists all ManagedClusterAddOns in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ManagedClusterAddOn, err error)
	// ManagedClusterAddOns returns an object that can list and get ManagedClusterAddOns.
	ManagedClusterAddOns(namespace string) ManagedClusterAddOnNamespaceLister
	ManagedClusterAddOnListerExpansion
}

// managedClusterAddOnLister implements the ManagedClusterAddOnLister interface.
type managedClusterAddOnLister struct {
	indexer cache.Indexer
}

// NewManagedClusterAddOnLister returns a new ManagedClusterAddOnLister.
func NewManagedClusterAddOnLister(indexer cache.Indexer) ManagedClusterAddOnLister {
	return &managedClusterAddOnLister{indexer: indexer}
}

// List lists all ManagedClusterAddOns in the indexer.
func (s *managedClusterAddOnLister) List(selector labels.Selector) (ret []*v1alpha1.ManagedClusterAddOn, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ManagedClusterAddOn))
	})
	return ret, err
}

// ManagedClusterAddOns returns an object that can list and get ManagedClusterAddOns.
func (s *managedClusterAddOnLister) ManagedClusterAddOns(namespace string) ManagedClusterAddOnNamespaceLister {
	return managedClusterAddOnNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ManagedClusterAddOnNamespaceLister helps list and get ManagedClusterAddOns.
// All objects returned here must be treated as read-only.
type ManagedClusterAddOnNamespaceLister interface {
	// List lists all ManagedClusterAddOns in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ManagedClusterAddOn, err error)
	// Get retrieves the ManagedClusterAddOn from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ManagedClusterAddOn, error)
	ManagedClusterAddOnNamespaceListerExpansion
}

// managedClusterAddOnNamespaceLister implements the ManagedClusterAddOnNamespaceLister
// interface.
type managedClusterAddOnNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ManagedClusterAddOns in the indexer for a given namespace.
func (s managedClusterAddOnNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ManagedClusterAddOn, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ManagedClusterAddOn))
	})
	return ret, err
}

// Get retrieves the ManagedClusterAddOn from the indexer for a given namespace and name.
func (s managedClusterAddOnNamespaceLister) Get(name string) (*v1alpha1.ManagedClusterAddOn, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("managedclusteraddon"), name)
	}
	return obj.(*v1alpha1.ManagedClusterAddOn), nil
}