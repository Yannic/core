// Copyright 2019 The Google Cloud Robotics Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/googlecloudrobotics/core/src/go/pkg/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AppLister helps list Apps.
type AppLister interface {
	// List lists all Apps in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.App, err error)
	// Get retrieves the App from the index for a given name.
	Get(name string) (*v1alpha1.App, error)
	AppListerExpansion
}

// appLister implements the AppLister interface.
type appLister struct {
	indexer cache.Indexer
}

// NewAppLister returns a new AppLister.
func NewAppLister(indexer cache.Indexer) AppLister {
	return &appLister{indexer: indexer}
}

// List lists all Apps in the indexer.
func (s *appLister) List(selector labels.Selector) (ret []*v1alpha1.App, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.App))
	})
	return ret, err
}

// Get retrieves the App from the index for a given name.
func (s *appLister) Get(name string) (*v1alpha1.App, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("app"), name)
	}
	return obj.(*v1alpha1.App), nil
}