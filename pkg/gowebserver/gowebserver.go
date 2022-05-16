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

package gowebserver

import (
	"fmt"
	"os"

	"github.com/jeremyje/gowebserver/pkg/certtool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Run() {
	_, syncFunc := configLogger(true)
	defer syncFunc()
	if err := platformMain(); err != nil {
		zap.S().Error(err)
		zap.S().Sync()
	}
}

func configLogger(verbose bool) (*zap.Logger, func() error) {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.Encoding = "console"
	if verbose {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err == nil {
		zap.ReplaceGlobals(logger)
	}
	if logger == nil {
		return nil, func() error { return nil }
	}
	return logger, logger.Sync
}

func runInteractive() error {
	terminateCh := make(chan error, 1)
	defer close(terminateCh)
	return runApplication(terminateCh)
}

func runApplication(termCh <-chan error) error {
	conf, err := Load()
	if err != nil {
		return err
	}
	logger, syncFunc := configLogger(conf.Verbose)
	defer syncFunc()

	logger.Sugar().Debug(conf)

	checkError(createCertificate(conf))

	httpServer, err := New(conf)
	if err != nil {
		return err
	}

	return httpServer.Serve(termCh)
}

func createCertificate(conf *Config) error {
	dir, err := os.Getwd()
	zap.S().With("certificate", conf.HTTPS.Certificate, "directory", dir, "error", err).Debug("createCertificate")
	_, certErr := os.Stat(conf.HTTPS.Certificate.CertificateFilePath)
	_, privateKeyErr := os.Stat(conf.HTTPS.Certificate.PrivateKeyFilePath)
	if conf.HTTPS.Certificate.ForceOverwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		var parentKP *certtool.KeyPair

		if conf.HTTPS.Certificate.RootCertificateFilePath != "" {
			rootCertPath := conf.HTTPS.Certificate.RootCertificateFilePath
			rootKeyPath := conf.HTTPS.Certificate.RootPrivateKeyFilePath
			kp, err := certtool.GenerateAndWriteKeyPair(&certtool.Args{
				KeyType: &certtool.KeyType{
					Algorithm: "RSA",
					KeyLength: 2048,
				},
				Validity: conf.HTTPS.Certificate.CertificateValidDuration,
				CA:       true,
			},
				rootCertPath,
				rootKeyPath)
			zap.S().With("error", err, "certificateFile", rootCertPath, "privateKeyFile", rootKeyPath).Debug("GenerateAndWriteKeyPair")
			if err != nil {
				return fmt.Errorf("cannot write public certificate, %s", err)
			}

			parentKP = kp
		}

		_, err := certtool.GenerateAndWriteKeyPair(&certtool.Args{
			KeyType: &certtool.KeyType{
				Algorithm: "RSA",
				KeyLength: 2048,
			},
			Validity:      conf.HTTPS.Certificate.CertificateValidDuration,
			CA:            false,
			ParentKeyPair: parentKP,
		},
			conf.HTTPS.Certificate.CertificateFilePath,
			conf.HTTPS.Certificate.PrivateKeyFilePath)
		zap.S().With("error", err, "certificateFile", conf.HTTPS.Certificate.CertificateFilePath, "privateKeyFile", conf.HTTPS.Certificate.PrivateKeyFilePath).Debug("GenerateAndWriteKeyPair")
		if err != nil {
			return fmt.Errorf("cannot write public certificate, %s", err)
		}
	}
	return nil
}
