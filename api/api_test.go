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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gocraft/web"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics-ng/tap-ca/client"
	"github.com/trustedanalytics-ng/tap-ca/engine"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const (
	UserName     string = "user"
	UserPassword string = "password"
)

func prepareMocksAndRouter(t *testing.T) (router *web.Router, mockEngineApi *engine.MockEngineApi, mockCtrl *gomock.Controller) {
	mockCtrl = gomock.NewController(t)
	mockEngineApi = engine.NewMockEngineApi(mockCtrl)

	Config = &ApiConfig{
		EngineApi: mockEngineApi,
	}

	router = SetupRouter(true)
	return
}

func getCaProperBasicAuth(router *web.Router, t *testing.T) *client.TapCaApiConnector {
	os.Setenv("USER", UserName)
	os.Setenv("PASS", UserPassword)

	testServer := httptest.NewServer(router)
	client, err := client.NewTapCaApiConnector(testServer.URL, UserName, UserPassword)
	if err != nil {
		t.Fatal("Ca client error: ", err)
	}
	return client
}

func TestGetCa(t *testing.T) {
	router, mockEngineApi, mockCtrl := prepareMocksAndRouter(t)
	client := getCaProperBasicAuth(router, t)

	certificateContent := "u2zZJWtxSvRagBp1wJ6M"
	certificateHash := "f124a60a"

	Convey(fmt.Sprintf("Given certificate: %s and hash of this certificate: %s", certificateContent, certificateHash),
		t, func() {
			Convey("GetCaCert should return proper response", func() {
				gomock.InOrder(
					mockEngineApi.EXPECT().GetCaCert().Return(certificateContent, nil),
					mockEngineApi.EXPECT().GetHashOfCaCert().Return(certificateHash, nil),
				)
				response, err := client.GetCa()

				So(err, ShouldBeNil)
				So(response.CaCertificateContent, ShouldEqual, certificateContent)
				So(response.Hash, ShouldEqual, certificateHash)

			})
			Reset(func() {
				mockCtrl.Finish()
			})
		})
}

func TestGetCertKey(t *testing.T) {
	router, mockEngineApi, mockCtrl := prepareMocksAndRouter(t)
	client := getCaProperBasicAuth(router, t)

	certContent := "0j7TVGQvGLsTr7yyQFKd"
	keyContent := "ZiT4botHmKFY84Cv6vxF"

	Convey(fmt.Sprintf("Given certificate %s and key %s", certContent, keyContent), t, func() {
		Convey("GenerateServiceKeys should return both", func() {
			mockEngineApi.EXPECT().GenerateServiceKeys("hostname").Return(certContent, keyContent, nil)

			response, err := client.GetCertKey("hostname")

			So(err, ShouldBeNil)
			So(response.CertificateContent, ShouldEqual, certContent)
			So(response.KeyContent, ShouldEqual, keyContent)
		})
		Reset(func() {
			mockCtrl.Finish()
		})
	})

	Convey("Should return error and Not Found message response for empty subhostname", t, func() {
		_, err := client.GetCertKey("")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Not Found")
	})

	Convey("Should fail properly if something goes wrong", t, func() {
		mockEngineApi.EXPECT().GenerateServiceKeys("hostname").Return("", "", errors.New("critical error!"))
		_, err := client.GetCertKey("hostname")

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Bad response")
		mockCtrl.Finish()
	})

}

func TestGetCaBundle(t *testing.T) {
	router, mockEngineApi, mockCtrl := prepareMocksAndRouter(t)
	client := getCaProperBasicAuth(router, t)

	bundleContent := "u2zZJWtxSvRagBp1wJ6M"

	Convey("Given bundle "+bundleContent, t, func() {
		Convey("GetCaBundle should return proper response", func() {
			mockEngineApi.EXPECT().GetCaBundle().Return(bundleContent, nil)

			response, err := client.GetCaBundle()

			So(err, ShouldBeNil)
			So(response.CaBundleContent, ShouldEqual, bundleContent)
		})
		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestGetHealthz(t *testing.T) {
	r, _, _ := prepareMocksAndRouter(t)
	Convey("Should return proper response", t, func() {
		rr := commonHttp.SendRequest("GET", "/healthz", nil, r, t)
		commonHttp.AssertResponse(rr, "Health OK", http.StatusOK)
	})
}

func readAndAssertJson(rr *httptest.ResponseRecorder, retstruct interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), &retstruct)
	So(err, ShouldBeNil)
}
