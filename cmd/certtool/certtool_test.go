// Copyright 2022 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArgsFromFlags(t *testing.T) {
	args, err := argsFromFlags()

	if err != nil {
		t.Fatalf("got error, %s", err)
	}
	if args.CA {
		t.Errorf("args.CA = %t, want false", args.CA)
	}
}

func TestStringToKeyType(t *testing.T) {
	testCases := []struct {
		input         string
		wantAlgorithm string
		wantKeyLength int
	}{
		{input: "", wantAlgorithm: "RSA", wantKeyLength: 2048},
		{input: "RSA", wantAlgorithm: "RSA", wantKeyLength: 2048},
		{input: "RSA-2048", wantAlgorithm: "RSA", wantKeyLength: 2048},
		{input: "rsa-4096", wantAlgorithm: "RSA", wantKeyLength: 4096},
		{input: "ecdsa-384", wantAlgorithm: "ECDSA", wantKeyLength: 384},
		{input: "ECDSA-521", wantAlgorithm: "ECDSA", wantKeyLength: 521},
		{input: "ECDSA", wantAlgorithm: "ECDSA", wantKeyLength: 521},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			gotAlgorithm, gotKeyLength, err := StringToKeyType(tc.input)
			if err != nil {
				t.Fatalf("got error, %s", err)
			}
			if tc.wantAlgorithm != gotAlgorithm {
				t.Errorf("algorithm want: %v, got: %v", tc.wantAlgorithm, gotAlgorithm)
			}
			if tc.wantKeyLength != gotKeyLength {
				t.Errorf("keyLength want: %v, got: %v", tc.wantAlgorithm, gotKeyLength)
			}
		})
	}
}

func TestExpandHostnames(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		{input: "", want: []string{}},
		{input: "localhost", want: []string{"localhost"}},
		{input: "localhost,gowebserver,localhost", want: []string{"gowebserver", "localhost"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := ExpandHostnames(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ExpandHostnames() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
