// Copyright 2019 Jeremy Edwards
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

package server

import (
	"fmt"
	"os"

	"github.com/jeremyje/gowebserver/cert"
	"github.com/jeremyje/gowebserver/config"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func platformMain() error {
	runAsService, err := svc.IsWindowsService()
	if err != nil {
		log.Printf("cannot determine if runnning as a service, %v, assuming standalone", err)
		runAsService = false
	}

	if runAsService {
		return runService()
	}

	return runInteractive()
}

// https://pkg.go.dev/golang.org/x/sys/windows/svc#Handler
type serviceHandler struct {
}

func (m *serviceHandler) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	changes <- svc.Status{State: svc.StartPending}

	resultCh := make(chan error, 1)
	terminateCh := make(chan error, 1)

	defer func() {
		close(terminateCh)
		close(resultCh)
	}()

	go func() {
		err := runInteractive(terminateCh)
		resultCh <- err
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	for {
		select {
		case err := <-resultCh:
			if err != nil {
				elog.Info(1, fmt.Sprintf("service died because %v", err))
				return true, util.StringHashUint32(err.Error())
			}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				select {
				case terminateCh <- errors.New("service is stopping"):
				default:
				}
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}

func runService(app *Application, isDebug bool) error {
	var err error
	if isDebug {
		elog = debug.New(app.Name)
	} else {
		elog, err = eventlog.Open(app.Name)
		if err != nil {
			return err
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", app.Name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(app.Name, &serviceHandler{
		app: app,
	})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", app.Name, err))
		return err
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", app.Name))
	return nil
}
