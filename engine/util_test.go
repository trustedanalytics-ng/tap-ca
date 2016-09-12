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
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFillTemplate(t *testing.T) {
	Convey("Test fillTemplate", t, func() {
		Convey("Should return proper result from template", func() {

			values := TemplateValues{"default-org", "default-common-name", ""}

			err := fillTemplate("../testData/templates/template.json", values, "../testData/templates/realResult.json")
			So(err, ShouldBeNil)

			realResult, err := ioutil.ReadFile("../testData/templates/realResult.json")
			So(err, ShouldBeNil)

			expectedResult, err := ioutil.ReadFile("../testData/templates/expectedResult.json")
			So(err, ShouldBeNil)

			So(string(realResult), ShouldEqual, string(expectedResult))

			os.Remove("../testData/templates/realResult.json")
		})
	})
}

func TestExecuteCommand(t *testing.T) {
	Convey("Test executeCommand", t, func() {
		Convey("Should proper execute command with stdout and err redirection", func() {

			outputContent := "some output"

			command := exec.Command("echo", outputContent)

			output, errorOutput, err := executeCommand(command)
			So(err, ShouldBeNil)
			So(errorOutput.String(), ShouldEqual, "")
			So(output.String(), ShouldEqual, outputContent+"\n")
		})
	})
}
