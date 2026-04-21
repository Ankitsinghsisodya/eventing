/*
Copyright 2024 The Knative Authors

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

package utils

import (
	"testing"
)

func TestSetKlogVerbosityFromConfigMap(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name:    "key absent",
			data:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "zero value",
			data:    map[string]string{KlogVerbosityKey: "0"},
			wantErr: false,
		},
		{
			name:    "empty value",
			data:    map[string]string{KlogVerbosityKey: ""},
			wantErr: false,
		},
		{
			name:    "valid level 5",
			data:    map[string]string{KlogVerbosityKey: "5"},
			wantErr: false,
		},
		{
			name:    "valid level 9",
			data:    map[string]string{KlogVerbosityKey: "9"},
			wantErr: false,
		},
		{
			name:    "invalid non-integer",
			data:    map[string]string{KlogVerbosityKey: "high"},
			wantErr: true,
		},
		{
			name:    "invalid float",
			data:    map[string]string{KlogVerbosityKey: "3.5"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetKlogVerbosityFromConfigMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetKlogVerbosityFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
