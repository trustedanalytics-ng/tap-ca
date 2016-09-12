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
package main

import (
	"github.com/trustedanalytics-ng/tap-ca/api"
	"github.com/trustedanalytics-ng/tap-ca/engine"
	httpGoCommon "github.com/trustedanalytics-ng/tap-go-common/http"
	loggerGoCommon "github.com/trustedanalytics-ng/tap-go-common/logger"
)

var logger, _ = loggerGoCommon.InitLogger("main")

func main() {

	err := engine.InitializeCertificates()
	if err != nil {
		logger.Fatal("Failed to initialize certs! Error: " + err.Error())
	}

	api.Config = &api.ApiConfig{
		EngineApi: &engine.EngineApiContext{},
	}

	r := api.SetupRouter(true)
	httpGoCommon.StartServer(r)
}
