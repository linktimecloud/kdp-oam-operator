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

var (
	ErrWebTerminalNotFound = NewExceptCode(404, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700404,
			Description: "web terminal not found, please check your specified def type",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	ErrTerminalNotFound = NewExceptCode(200, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700001,
			Description: "web terminal not found",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	// ErrCreateTerminalFailed
	ErrCreateTerminalFailed = NewExceptCode(200, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700101,
			Description: "web terminal create failed",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	ErrObtainLimitTry = NewExceptCode(200, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700201,
			Description: "The number of attempts to obtain terminal information exceeded the limit. please try again",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	ErrIngressCheckFailed = NewExceptCode(200, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700301,
			Description: "web terminal url check failed, please try again",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	ErrServiceNotFound = NewExceptCode(200, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        700401,
			Description: "web terminal service check failed.",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})
)
