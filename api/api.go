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
package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-ca/engine"
	"github.com/trustedanalytics-ng/tap-ca/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

const (
	apiPrefix = "/api/v1"
)

type ApiConfig struct {
	EngineApi engine.EngineApi
}
type Context struct{}

var Config *ApiConfig
var logger, _ = commonLogger.InitLogger("api")

func SetupRouter(middlewareActivated bool) *web.Router {
	context := Context{}
	r := web.New(context)

	r.Get("/healthz", context.Healthz)

	basicAuthRouter := r.Subrouter(context, apiPrefix)

	if middlewareActivated {
		basicAuthRouter.Middleware(web.LoggerMiddleware)
		basicAuthRouter.Middleware(context.BasicAuthorizeMiddleware)
	}

	basicAuthRouter.Get("/ca", context.Ca)
	basicAuthRouter.Get("/certkey/:subhostname", context.CertKey)
	basicAuthRouter.Get("/ca-bundle", context.CaBundle)

	return r
}

func (c *Context) Healthz(rw web.ResponseWriter, req *web.Request) {
	fmt.Fprint(rw, "Health OK")
}

func (c *Context) Ca(rw web.ResponseWriter, req *web.Request) {
	ca_pem, err := Config.EngineApi.GetCaCert()
	if err != nil {
		logger.Error(err)
		commonHttp.Respond500(rw, errors.New("cannot get ca certificate!"))
		return
	}

	hash, err := Config.EngineApi.GetHashOfCaCert()
	if err != nil {
		logger.Error(err)
		commonHttp.Respond500(rw, errors.New("couldn't get ca certificate's hash!"))
		return
	}

	commonHttp.WriteJson(rw, models.CaResponse{CaCertificateContent: ca_pem, Hash: hash}, http.StatusOK)
}

func (c *Context) CertKey(rw web.ResponseWriter, req *web.Request) {
	name := req.PathParams["subhostname"]

	cert, key, err := Config.EngineApi.GenerateServiceKeys(name)
	if err != nil {
		logger.Error(err)
		commonHttp.Respond500(rw, errors.New("cannot get certificate and key for service!"))
		return
	}
	commonHttp.WriteJson(rw, models.CertKeyResponse{CertificateContent: cert, KeyContent: key}, http.StatusOK)
}

func (c *Context) CaBundle(rw web.ResponseWriter, req *web.Request) {
	caBundle, err := Config.EngineApi.GetCaBundle()
	if err != nil {
		logger.Error(err)
		commonHttp.Respond500(rw, errors.New("cannot get ca certificates bundle!"))
		return
	}

	commonHttp.WriteJson(rw, models.CaBundleResponse{CaBundleContent: caBundle}, http.StatusOK)
}
