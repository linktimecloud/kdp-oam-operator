/*
Copyright 2023 KDP(Kubernetes Data Platform).

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

package uuid

import (
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

func GenAppUUID(namespace string, name string, length int) string {
	return genAppUUID(uuid5(namespace, name), length)
}

// AppUuid generates a UUID from a namespace and a name.Version 5, based on SHA-1 hashing (RFC 4122)
func genAppUUID(hashStr string, length int) string {
	hasher := sha512.New()
	hasher.Write([]byte(hashStr))
	hashedValue := hasher.Sum(nil)
	hashStr = fmt.Sprintf("%x", hashedValue)[:8]
	hashInt, _ := strconv.ParseInt(hashStr, 16, 64)
	b64Data := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(hashInt))))
	b64Data = strings.Replace(b64Data, "+", "", -1)
	b64Data = strings.Replace(b64Data, "/", "", -1)
	b64Data = strings.Replace(b64Data, "=", "", -1)

	if len(b64Data) > length {
		b64Data = b64Data[:length]
	}

	return strings.ToLower(b64Data)
}

// 这个函数首先创建一个新的SHA-1 hash，然后写入命名空间和名称，然后对结果进行哈希。
// 然后，它复制哈希的前16个字节到返回的UUID中，并且设置UUID的版本为5 (表示使用SHA-1哈希和命名空间的UUID)以及变体为10（表示此UUID是基于RFC 4122规范）。
func uuid5(namespace, name string) string {
	var uuid [16]byte
	h := sha1.New()
	h.Write([]byte(namespace))
	h.Write([]byte(name))
	hash := h.Sum(nil)
	copy(uuid[:], hash)
	uuid[6] = (uuid[6] & 0x0f) | 0x50 // Version 5
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%x", uuid)
}
