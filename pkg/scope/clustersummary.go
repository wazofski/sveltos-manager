/*
Copyright 2022. projectsveltos.io. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scope

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/projectsveltos/sveltos-manager/api/v1alpha1"
)

// ClusterSummaryScopeParams defines the input parameters used to create a new ClusterSummary Scope.
type ClusterSummaryScopeParams struct {
	Client         client.Client
	Logger         logr.Logger
	ClusterProfile *configv1alpha1.ClusterProfile
	ClusterSummary *configv1alpha1.ClusterSummary
	ControllerName string
}

// NewClusterSummaryScope creates a new ClusterSummary Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterSummaryScope(params ClusterSummaryScopeParams) (*ClusterSummaryScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a ClusterSummaryScope")
	}
	if params.ClusterSummary == nil {
		return nil, errors.New("failed to generate new scope from nil ClusterSummary")
	}

	helper, err := patch.NewHelper(params.ClusterSummary, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}
	return &ClusterSummaryScope{
		Logger:         params.Logger,
		client:         params.Client,
		ClusterProfile: params.ClusterProfile,
		ClusterSummary: params.ClusterSummary,
		patchHelper:    helper,
		controllerName: params.ControllerName,
	}, nil
}

// ClusterSummaryScope defines the basic context for an actuator to operate upon.
type ClusterSummaryScope struct {
	logr.Logger
	client         client.Client
	patchHelper    *patch.Helper
	ClusterProfile *configv1alpha1.ClusterProfile
	ClusterSummary *configv1alpha1.ClusterSummary
	controllerName string
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterSummaryScope) PatchObject(ctx context.Context) error {
	return s.patchHelper.Patch(
		ctx,
		s.ClusterSummary,
	)
}

// Close closes the current scope persisting the clusterprofile configuration and status.
func (s *ClusterSummaryScope) Close(ctx context.Context) error {
	return s.PatchObject(ctx)
}

// Name returns the ClusterSummary name.
func (s *ClusterSummaryScope) Name() string {
	return s.ClusterSummary.Name
}

// Namespace returns the ClusterSummary namespace.
func (s *ClusterSummaryScope) Namespace() string {
	return s.ClusterSummary.Namespace
}

func (s *ClusterSummaryScope) initializeFeatureStatusSummary() {
	if s.ClusterSummary.Status.FeatureSummaries == nil {
		s.ClusterSummary.Status.FeatureSummaries = make([]configv1alpha1.FeatureSummary, 0)
	}
}

// SetFeatureStatus sets the feature status.
func (s *ClusterSummaryScope) SetFeatureStatus(featureID configv1alpha1.FeatureID,
	status configv1alpha1.FeatureStatus, hash []byte) {

	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.ClusterSummary.Status.FeatureSummaries[i].Status = status
			s.ClusterSummary.Status.FeatureSummaries[i].Hash = hash
			return
		}
	}

	s.initializeFeatureStatusSummary()

	s.ClusterSummary.Status.FeatureSummaries = append(
		s.ClusterSummary.Status.FeatureSummaries,
		configv1alpha1.FeatureSummary{
			FeatureID: featureID,
			Status:    status,
			Hash:      hash,
		},
	)
}

// SetFailureMessage sets the infrastructure status failure message.
func (s *ClusterSummaryScope) SetFailureMessage(featureID configv1alpha1.FeatureID, failureMessage *string) {
	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.ClusterSummary.Status.FeatureSummaries[i].FailureMessage = failureMessage
			return
		}
	}

	s.initializeFeatureStatusSummary()

	s.ClusterSummary.Status.FeatureSummaries = append(
		s.ClusterSummary.Status.FeatureSummaries,
		configv1alpha1.FeatureSummary{
			FeatureID:      featureID,
			FailureMessage: failureMessage,
		},
	)
}

// SetFailureReason sets the feature status failure reason.
func (s *ClusterSummaryScope) SetFailureReason(featureID configv1alpha1.FeatureID,
	failureReason *string) {

	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.ClusterSummary.Status.FeatureSummaries[i].FailureReason = failureReason
			return
		}
	}

	s.initializeFeatureStatusSummary()

	s.ClusterSummary.Status.FeatureSummaries = append(
		s.ClusterSummary.Status.FeatureSummaries,
		configv1alpha1.FeatureSummary{
			FeatureID:     featureID,
			FailureReason: failureReason,
		},
	)
}

// SetDeployedGroupVersionKind sets the list of deployed GroupVersionKinds
func (s *ClusterSummaryScope) SetDeployedGroupVersionKind(featureID configv1alpha1.FeatureID,
	deployed []schema.GroupVersionKind) {

	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.updateDeployedGroupVersionKind(&s.ClusterSummary.Status.FeatureSummaries[i], deployed)
			return
		}
	}

	s.initializeFeatureStatusSummary()

	s.ClusterSummary.Status.FeatureSummaries = append(
		s.ClusterSummary.Status.FeatureSummaries,
		configv1alpha1.FeatureSummary{
			FeatureID: featureID,
		},
	)

	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.updateDeployedGroupVersionKind(&s.ClusterSummary.Status.FeatureSummaries[i], deployed)
			return
		}
	}
}

func (s *ClusterSummaryScope) updateDeployedGroupVersionKind(fs *configv1alpha1.FeatureSummary,
	deployed []schema.GroupVersionKind) {

	tmp := make([]string, 0)

	// Preserve the order
	current := make(map[string]bool)
	for _, k := range fs.DeployedGroupVersionKind {
		current[k] = true
		tmp = append(tmp, k)
	}

	for i := range deployed {
		key := fmt.Sprintf("%s.%s.%s", deployed[i].Kind, deployed[i].Version, deployed[i].Group)
		if _, ok := current[key]; !ok {
			current[key] = true
			tmp = append(tmp, key)
		}
	}

	fs.DeployedGroupVersionKind = tmp
}

func (s *ClusterSummaryScope) SetLastAppliedTime(featureID configv1alpha1.FeatureID,
	lastAppliedTime *metav1.Time) {

	for i := range s.ClusterSummary.Status.FeatureSummaries {
		if s.ClusterSummary.Status.FeatureSummaries[i].FeatureID == featureID {
			s.ClusterSummary.Status.FeatureSummaries[i].LastAppliedTime = lastAppliedTime
			return
		}
	}

	s.initializeFeatureStatusSummary()

	s.ClusterSummary.Status.FeatureSummaries = append(
		s.ClusterSummary.Status.FeatureSummaries,
		configv1alpha1.FeatureSummary{
			FeatureID:       featureID,
			LastAppliedTime: lastAppliedTime,
		},
	)
}

// IsContinuousSync returns true if ClusterProfile is set to keep updating workload cluster
func (s *ClusterSummaryScope) IsContinuousSync() bool {
	return s.ClusterSummary.Spec.ClusterProfileSpec.SyncMode == configv1alpha1.SyncModeContinuous ||
		s.ClusterSummary.Spec.ClusterProfileSpec.SyncMode == configv1alpha1.SyncModeContinuousWithDriftDetection
}

// IsOneTimeSync returns true if ClusterProfile sync mod is set to one time
func (s *ClusterSummaryScope) IsOneTimeSync() bool {
	return s.ClusterSummary.Spec.ClusterProfileSpec.SyncMode == configv1alpha1.SyncModeOneTime
}

// IsDryRunSync returns true if ClusterProfile sync mod is set to dryRun
func (s *ClusterSummaryScope) IsDryRunSync() bool {
	return s.ClusterSummary.Spec.ClusterProfileSpec.SyncMode == configv1alpha1.SyncModeDryRun
}
