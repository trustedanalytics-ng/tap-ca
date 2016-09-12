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
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
)

func getEnvOrDefault(envName string, defaultValue string) string {
	value := os.Getenv(envName)
	if value == "" {
		return defaultValue
	}
	return value
}

func executeCommand(command *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return stdout, stderr, err
}

type TemplateFiller struct {
	TemplatePath string
	OutputPath   string
}

func fillTemplateWithArray(paths []TemplateFiller, valuesToFillWith TemplateValues) error {
	for _, path := range paths {
		err := fillTemplate(path.TemplatePath, valuesToFillWith, path.OutputPath)
		if err != nil {
			return err
		}
	}
	return nil
}

type FileWriter struct {
	Path string
	Data []byte
}

func writeFileWithArray(fileWrites []FileWriter, filePerm os.FileMode) error {
	for _, fileWrite := range fileWrites {
		err := ioutil.WriteFile(fileWrite.Path, fileWrite.Data, filePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillTemplate(templatePath string, valuesToFillWith TemplateValues, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = fillTemplateGeneric(templatePath, valuesToFillWith, f)
	return err
}

func fillTemplateTempfile(templatePath string, valuesToFillWith TemplateValues) (string, error) {
	tempfile, err := ioutil.TempFile("", "ca-req")
	if err != nil {
		return "", err
	}
	defer tempfile.Close()

	err = fillTemplateGeneric(templatePath, valuesToFillWith, tempfile)
	return tempfile.Name(), nil
}

func fillTemplateGeneric(templatePath string, valuesToFillWith TemplateValues, outFile *os.File) error {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	err = t.Execute(outFile, valuesToFillWith)
	return err
}
