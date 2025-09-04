/*
Copyright 2021.

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

package vaultresourcecontroller

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// MockConditionsAware is a mock implementation using testify/mock
type MockConditionsAware struct {
	mock.Mock
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (m *MockConditionsAware) GetConditions() []metav1.Condition {
	args := m.Called()
	return args.Get(0).([]metav1.Condition)
}

func (m *MockConditionsAware) SetConditions(conditions []metav1.Condition) {
	m.Called(conditions)
}

func (m *MockConditionsAware) GetObjectKind() schema.ObjectKind {
	return &m.TypeMeta
}

func (m *MockConditionsAware) DeepCopyObject() runtime.Object {
	args := m.Called()
	if obj := args.Get(0); obj != nil {
		return obj.(runtime.Object)
	}
	return nil
}

func (m *MockConditionsAware) DeepCopyInto(out *MockConditionsAware) {
	m.Called(out)
}

// NewMockConditionsAware creates a new mock with common setup
func NewMockConditionsAware(generation int64, conditions []metav1.Condition) *MockConditionsAware {
	mock := &MockConditionsAware{
		ObjectMeta: metav1.ObjectMeta{
			Generation: generation,
		},
	}
	if conditions != nil {
		mock.On("GetConditions").Return(conditions)
	}
	return mock
}

func TestIsDriftDetectionEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "Disabled when env var not set",
			envValue: "",
			expected: false,
		},
		{
			name:     "Enabled when env var is true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "Disabled when env var is false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "Disabled when env var is invalid",
			envValue: "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv("ENABLE_DRIFT_DETECTION", tt.envValue)
			} else {
				os.Unsetenv("ENABLE_DRIFT_DETECTION")
			}
			defer os.Unsetenv("ENABLE_DRIFT_DETECTION")

			result := IsDriftDetectionEnabled()
			if result != tt.expected {
				t.Errorf("IsDriftDetectionEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPeriodicReconcilePredicate_Update(t *testing.T) {
	predicate := NewPeriodicReconcilePredicate(5 * time.Minute)

	tests := []struct {
		name                  string
		oldGeneration         int64
		newGeneration         int64
		conditions            []metav1.Condition
		driftDetectionEnabled bool
		expectedResult        bool
		description           string
	}{
		{
			name:                  "Should reconcile when generation changes regardless of drift detection setting",
			oldGeneration:         1,
			newGeneration:         2,
			conditions:            []metav1.Condition{},
			driftDetectionEnabled: false,
			expectedResult:        true,
			description:           "Generation change should always trigger reconcile",
		},
		{
			name:          "Should reconcile when interval has elapsed and drift detection is enabled",
			oldGeneration: 1,
			newGeneration: 1,
			conditions: []metav1.Condition{
				{
					Type:               ReconcileSuccessful,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)), // 10 minutes ago
				},
			},
			driftDetectionEnabled: true,
			expectedResult:        true,
			description:           "Should reconcile when enough time has passed and drift detection is enabled",
		},
		{
			name:          "Should not reconcile when interval has elapsed but drift detection is disabled",
			oldGeneration: 1,
			newGeneration: 1,
			conditions: []metav1.Condition{
				{
					Type:               ReconcileSuccessful,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)), // 10 minutes ago
				},
			},
			driftDetectionEnabled: false,
			expectedResult:        false,
			description:           "Should not reconcile when drift detection is disabled even if time has elapsed",
		},
		{
			name:          "Should not reconcile when interval has not elapsed",
			oldGeneration: 1,
			newGeneration: 1,
			conditions: []metav1.Condition{
				{
					Type:               ReconcileSuccessful,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-2 * time.Minute)), // 2 minutes ago
				},
			},
			driftDetectionEnabled: true,
			expectedResult:        false,
			description:           "Should not reconcile when not enough time has passed",
		},
		{
			name:          "Should not reconcile when last reconcile failed",
			oldGeneration: 1,
			newGeneration: 1,
			conditions: []metav1.Condition{
				{
					Type:               ReconcileFailed,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
				},
			},
			driftDetectionEnabled: true,
			expectedResult:        false,
			description:           "Should not reconcile based on time when last reconcile failed",
		},
		{
			name:                  "Should not reconcile when no conditions exist and drift detection is enabled",
			oldGeneration:         1,
			newGeneration:         1,
			conditions:            []metav1.Condition{},
			driftDetectionEnabled: true,
			expectedResult:        false,
			description:           "Should not reconcile when no conditions exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up drift detection environment
			if tt.driftDetectionEnabled {
				os.Setenv("ENABLE_DRIFT_DETECTION", "true")
			} else {
				os.Unsetenv("ENABLE_DRIFT_DETECTION")
			}
			defer os.Unsetenv("ENABLE_DRIFT_DETECTION")

			// Create mock objects
			oldObj := &MockConditionsAware{
				ObjectMeta: metav1.ObjectMeta{
					Generation: tt.oldGeneration,
				},
			}
			newObj := &MockConditionsAware{
				ObjectMeta: metav1.ObjectMeta{
					Generation: tt.newGeneration,
				},
			}

			// Set up mock expectations for GetConditions only when needed
			// GetConditions is only called when:
			// 1. Generation hasn't changed AND
			// 2. Drift detection is enabled
			if tt.oldGeneration == tt.newGeneration && tt.driftDetectionEnabled {
				newObj.On("GetConditions").Return(tt.conditions)
			}

			// Create update event
			updateEvent := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}

			// Test the predicate
			result := predicate.Update(updateEvent)
			if result != tt.expectedResult {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.expectedResult, result)
			}

			// Assert that all expectations were met
			newObj.AssertExpectations(t)
		})
	}
}
