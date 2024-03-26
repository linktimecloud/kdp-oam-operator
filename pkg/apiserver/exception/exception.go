/*
Copyright 2024 KDP(Kubernetes Data Platform).

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exception

import (
	"errors"
	"fmt"

	"kdp-oam-operator/pkg/apiserver/utils"
	"kdp-oam-operator/pkg/utils/log"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-playground/validator/v10"
)

// ExceptCode exception code
type ExceptCode struct {
	ErrDetail     ErrDetail `json:"error"`
	HTTPCode      int32     `json:"-"`
	ExceptionCode int32     `json:"-"`
	Status        int32     `json:"status"`
	Message       string    `json:"message"`
}

type ErrDetail struct {
	AppName        string  `json:"app"`
	ExceptionLevel string  `json:"type"`
	ErrInfo        ErrInfo `json:"info"`
}

type ErrInfo struct {
	Code        int32  `json:"code"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
	ManualURL   string `json:"manual"`
}

func (e *ExceptCode) Error() string {
	return fmt.Sprintf("HTTPCode:%d ExceptionCode:%d Message:%s", e.HTTPCode, e.ExceptionCode, e.Message)
}

var exceptCodeMap map[int32]*ExceptCode

// NewExceptCode new exception code
func NewExceptCode(httpCode int32, errDetail ErrDetail) *ExceptCode {
	exceptionCode := errDetail.ErrInfo.Code
	message := errDetail.ErrInfo.Description
	if exceptCodeMap == nil {
		exceptCodeMap = make(map[int32]*ExceptCode)
	}
	if _, exit := exceptCodeMap[exceptionCode]; exit {
		panic("exception code is exist")
	}
	exceptCode := &ExceptCode{
		HTTPCode:      httpCode,
		ExceptionCode: exceptionCode,
		Message:       message,
		Status:        1,
		ErrDetail: ErrDetail{
			AppName:        "apiserver",
			ExceptionLevel: "error",
			ErrInfo: ErrInfo{
				Code:        exceptionCode,
				Description: message,
				Solution:    errDetail.ErrInfo.Solution,
				ManualURL:   errDetail.ErrInfo.ManualURL,
			},
		},
	}
	exceptCodeMap[exceptionCode] = exceptCode
	return exceptCode
}

// ReturnError Unified handling of all types of errors, generating a standard return structure.
func ReturnError(req *restful.Request, res *restful.Response, err error) {
	var exceptcode *ExceptCode
	if errors.As(err, &exceptcode) {
		if err := res.WriteHeaderAndEntity(int(exceptcode.HTTPCode), err); err != nil {
			log.Logger.Error("write entity failure %s", err.Error())
		}
		return
	}

	var restfulerr restful.ServiceError
	if errors.As(err, &restfulerr) {
		if err := res.WriteHeaderAndEntity(restfulerr.Code, ExceptCode{HTTPCode: int32(restfulerr.Code), ExceptionCode: int32(restfulerr.Code), Status: 1, Message: restfulerr.Message}); err != nil {
			log.Logger.Error("write entity failure %s", err.Error())
		}
		return
	}

	var validErr validator.ValidationErrors
	if errors.As(err, &validErr) {
		if err := res.WriteHeaderAndEntity(400, ExceptCode{HTTPCode: 500, ExceptionCode: 500, Status: 1, Message: err.Error()}); err != nil {
			log.Logger.Error("write entity failure %s", err.Error())
		}
		return
	}

	log.Logger.Errorf("Business exceptions, error message: %s, path:%s method:%s", err.Error(), utils.Sanitize(req.Request.URL.String()), req.Request.Method)
	if err := res.WriteHeaderAndEntity(500, ExceptCode{HTTPCode: 500, ExceptionCode: 500, Status: 1, Message: err.Error()}); err != nil {
		log.Logger.Error("write entity failure %s", err.Error())
	}
}
