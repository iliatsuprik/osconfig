//  Copyright 2026 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ospatch

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/GoogleCloudPlatform/osconfig/packages"
	utilmocks "github.com/GoogleCloudPlatform/osconfig/util/mocks"
	"github.com/GoogleCloudPlatform/osconfig/util/utiltest"
	"github.com/golang/mock/gomock"
)

func TestSystemRebootRequiredApt(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		desc       string
		setup      func(t *testing.T)
		wantReboot bool
		wantErr    error
	}{
		{
			desc: "no reboot required when reboot file does not exist",
			setup: func(t *testing.T) {
				utiltest.OverrideVariable(t, &packages.AptExists, true)
				utiltest.OverrideVariable(t, &rebootRequiredFile, "/non_existing_reboot_file")
			},
			wantReboot: false,
		},
		{
			desc: "reboot required when reboot file exists",
			setup: func(t *testing.T) {
				utiltest.OverrideVariable(t, &packages.AptExists, true)
				utiltest.OverrideVariable(t, &rebootRequiredFile, utiltest.CreateTempFile(t, "reboot-required"))
			},
			wantReboot: true,
		},
		{
			desc: "file read error is propagated",
			setup: func(t *testing.T) {
				utiltest.OverrideVariable(t, &packages.AptExists, true)
				utiltest.OverrideVariable(t, &rebootRequiredFile, "/dev/null/invalid")
			},
			wantReboot: false,
			wantErr:    &os.PathError{Op: "open", Path: "/dev/null/invalid", Err: syscall.ENOTDIR},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tt.setup(t)

			gotReboot, gotErr := SystemRebootRequired(ctx)

			utiltest.AssertEquals(t, gotReboot, tt.wantReboot)
			utiltest.AssertErrorMatch(t, gotErr, tt.wantErr)
		})
	}
}

func TestSystemRebootRequiredRpm(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	t.Cleanup(func() {
		mockCtrl.Finish()
	})

	tests := []struct {
		desc       string
		setup      func(t *testing.T)
		wantReboot bool
		wantErr    error
	}{
		{
			desc: "rpm reboot check succeeds",
			setup: func(t *testing.T) {
				utiltest.OverrideVariable(t, &packages.AptExists, false)
				utiltest.OverrideVariable(t, &rpmquery, utiltest.CreateTempFile(t, "rpmquery"))

				mockCommandRunner := utilmocks.NewMockCommandRunner(mockCtrl)
				runner = mockCommandRunner
				mockCommandRunner.EXPECT().Run(ctx, gomock.Any()).Return([]byte("1000\n"), nil, nil).Times(1)
			},
			wantReboot: false,
		},
		{
			desc: "unsupported package manager returns error",
			setup: func(t *testing.T) {
				utiltest.OverrideVariable(t, &packages.AptExists, false)
				utiltest.OverrideVariable(t, &rpmquery, "/non_existing_file")
			},
			wantErr: errors.New("no recognized package manager installed, can't determine if reboot is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tt.setup(t)

			gotReboot, gotErr := SystemRebootRequired(ctx)

			utiltest.AssertEquals(t, gotReboot, tt.wantReboot)
			utiltest.AssertErrorMatch(t, gotErr, tt.wantErr)
		})
	}
}

func TestInstallWUAUpdates(t *testing.T) {
	if err := InstallWUAUpdates(context.Background()); err != nil {
		t.Errorf("InstallWUAUpdates() on linux stub should not return an error, got: %v", err)
	}
}
