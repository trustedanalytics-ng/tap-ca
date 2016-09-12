/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	loggerGoCommon "github.com/trustedanalytics-ng/tap-go-common/logger"
)

const (
	dataDirPerm      = 0700
	filePerm         = 0644
	dataDirName      = "data"
	templatesDirName = "templates"
	caPemFile        = "ca.pem"
	caKeyFile        = "ca-key.pem"
	caCsrPemFile     = "ca-csr.pem"
	orgDefault       = "TAP"
	reqFile          = "req.json"
	tempReqFile      = "temp-req.json"
	caCsrFile        = "ca-csr.json"
	configFile       = "config.json"
	caBundleFile     = "ca-certificates.crt"
)

type EngineApi interface {
	GenerateServiceKeys(name string) (string, string, error)
	GetCaCert() (string, error)
	GetHashOfCaCert() (string, error)
	GetCaBundle() (string, error)
}

type EngineApiContext struct{}

type CertKey struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
	Csr  string `json:"csr"`
}

type TemplateValues struct {
	Org        string
	CommonName string
	CaCert     string
}

var logger, _ = loggerGoCommon.InitLogger("engine")
var org = getEnvOrDefault("ORG", orgDefault)

func InitializeCertificates() error {
	values := TemplateValues{Org: org, CommonName: "", CaCert: ""}

	err := os.Mkdir(dataDirName, dataDirPerm)
	if err != nil {
		if !os.IsExist(err) {
			logger.Error("cannot create data directory!")
			return err
		}
	}

	_, err = os.Stat(fmt.Sprintf("%s/%s", dataDirName, reqFile))
	if !(err != nil && os.IsNotExist(err)) {
		logger.Info("CA was already created.")
		return nil
	}

	templateFiller := []TemplateFiller{
		{
			TemplatePath: fmt.Sprintf("./%s/%s", templatesDirName, caCsrFile),
			OutputPath:   fmt.Sprintf("./%s/%s", dataDirName, caCsrFile),
		},
		{
			TemplatePath: fmt.Sprintf("./%s/%s", templatesDirName, configFile),
			OutputPath:   fmt.Sprintf("./%s/%s", dataDirName, configFile),
		},
		{
			TemplatePath: fmt.Sprintf("./%s/%s", templatesDirName, reqFile),
			OutputPath:   fmt.Sprintf("./%s/%s", dataDirName, reqFile),
		},
	}
	err = fillTemplateWithArray(templateFiller, values)
	if err != nil {
		return err
	}

	out, stderr, err := executeCommand(exec.Command(
		"./cfssl", "genkey", "-initca", fmt.Sprintf("%s/%s", dataDirName, caCsrFile)))
	if err != nil {
		logger.Error(stderr.String())
		return err
	}

	var certkey CertKey
	err = json.Unmarshal(out.Bytes(), &certkey)
	if err != nil {
		return err
	}

	filesWrite := []FileWriter{
		{
			Path: fmt.Sprintf("%s/%s", dataDirName, caPemFile),
			Data: []byte(certkey.Cert),
		},
		{
			Path: fmt.Sprintf("%s/%s", dataDirName, caKeyFile),
			Data: []byte(certkey.Key),
		},
		{
			Path: fmt.Sprintf("%s/%s", dataDirName, caCsrPemFile),
			Data: []byte(certkey.Csr),
		},
	}
	err = writeFileWithArray(filesWrite, filePerm)
	if err != nil {
		return err
	}

	err = fillTemplate(fmt.Sprintf("./%s/%s", templatesDirName, caBundleFile),
		TemplateValues{Org: org, CommonName: "", CaCert: certkey.Cert},
		fmt.Sprintf("./%s/%s", dataDirName, caBundleFile))
	if err != nil {
		return err
	}

	logger.Info("Initialize OK.")

	return nil
}

func (c *EngineApiContext) GenerateServiceKeys(name string) (string, string, error) {
	values := TemplateValues{Org: org, CommonName: name, CaCert: ""}

	filledTemplateFilepath, err := fillTemplateTempfile(fmt.Sprintf("./%s/%s", templatesDirName, reqFile), values)
	if err != nil {
		return "", "", err
	}
	defer os.Remove(filledTemplateFilepath)

	cmd := exec.Command("./cfssl",
		"gencert",
		"-loglevel=0",
		"-ca="+fmt.Sprintf("./%s/%s", dataDirName, caPemFile),
		"-ca-key="+fmt.Sprintf("./%s/%s", dataDirName, caKeyFile),
		"-config="+fmt.Sprintf("./%s/%s", dataDirName, configFile),
		"-profile=www",
		"-hostname="+name,
		filledTemplateFilepath)

	out, stderr, err := executeCommand(cmd)

	if err != nil {
		logger.Error("Stderr: ", stderr.String())
		return "", "", err
	}

	var certkey CertKey
	err = json.Unmarshal(out.Bytes(), &certkey)
	if err != nil {
		logger.Error("Unmarshal failed! output: " + out.String())
		return "", "", err
	}

	return certkey.Cert, certkey.Key, nil
}

func (c *EngineApiContext) GetCaCert() (string, error) {
	ca_pem_bytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dataDirName, caPemFile))
	if err != nil {
		return "", err
	}

	return string(ca_pem_bytes), nil
}

func (c *EngineApiContext) GetHashOfCaCert() (string, error) {
	cmd := exec.Command("/usr/bin/openssl", "x509", "-hash", "-noout", "-in", dataDirName+"/"+caPemFile)

	out, stderr, err := executeCommand(cmd)
	if err != nil {
		logger.Error("Stderr: ", stderr.String())
		return "", err
	}

	return strings.Trim(out.String(), "\n"), nil
}

func (c *EngineApiContext) GetCaBundle() (string, error) {
	ca_certificates_bytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dataDirName, caBundleFile))
	if err != nil {
		return "", err
	}

	return string(ca_certificates_bytes), nil
}
