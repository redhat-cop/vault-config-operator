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
		name           string
		oldGeneration  int64
		newGeneration  int64
		expectedResult bool
	}{
		{
			name:           "Should reconcile when generation changes",
			oldGeneration:  1,
			newGeneration:  2,
			expectedResult: true,
		},
		{
			name:           "Should not reconcile when generation is unchanged",
			oldGeneration:  1,
			newGeneration:  1,
			expectedResult: false,
		},
		{
			name:           "Should handle nil ObjectOld",
			oldGeneration:  0,
			newGeneration:  1,
			expectedResult: false,
		},
		{
			name:           "Should handle nil ObjectNew",
			oldGeneration:  1,
			newGeneration:  0,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updateEvent event.UpdateEvent

			if tt.name == "Should handle nil ObjectOld" {
				updateEvent = event.UpdateEvent{
					ObjectOld: nil,
					ObjectNew: &MockConditionsAware{
						ObjectMeta: metav1.ObjectMeta{Generation: tt.newGeneration},
					},
				}
			} else if tt.name == "Should handle nil ObjectNew" {
				updateEvent = event.UpdateEvent{
					ObjectOld: &MockConditionsAware{
						ObjectMeta: metav1.ObjectMeta{Generation: tt.oldGeneration},
					},
					ObjectNew: nil,
				}
			} else {
				updateEvent = event.UpdateEvent{
					ObjectOld: &MockConditionsAware{
						ObjectMeta: metav1.ObjectMeta{Generation: tt.oldGeneration},
					},
					ObjectNew: &MockConditionsAware{
						ObjectMeta: metav1.ObjectMeta{Generation: tt.newGeneration},
					},
				}
			}

			result := predicate.Update(updateEvent)
			if result != tt.expectedResult {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expectedResult, result)
			}
		})
	}
}
