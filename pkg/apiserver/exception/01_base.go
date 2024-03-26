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

// ErrServer an unexpected mistake.
var (
	ErrServer = NewExceptCode(500, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        100500,
			Description: "The service has lapsed.",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	// ErrForbidden check user perms failure
	ErrForbidden = NewExceptCode(403, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        100403,
			Description: "403 Forbidden.",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	// ErrUnauthorized check user auth failure
	ErrUnauthorized = NewExceptCode(401, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        100401,
			Description: "401 Unauthorized.",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})

	ErrInitKubeClient = NewExceptCode(500, ErrDetail{
		ExceptionLevel: "error",
		ErrInfo: ErrInfo{
			Code:        100402,
			Description: "Init kube client failure.",
			Solution:    "",
			ManualURL:   "",
		},
		AppName: "",
	})
)
