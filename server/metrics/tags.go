// Copyright 2022 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"errors"
	"strconv"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	api "github.com/tigrisdata/tigris/api/server/v1"
)

func mergeTags(tagSets ...map[string]string) map[string]string {
	res := make(map[string]string)
	for _, tagSet := range tagSets {
		for k, v := range tagSet {
			if _, ok := res[k]; !ok {
				res[k] = v
			} else {
				if res[k] == "unknown" {
					res[k] = v
				}
			}
		}
	}
	return res
}

func getFdbError(err error) (string, bool) {
	var fdbErr fdb.Error
	if errors.As(err, &fdbErr) {
		return strconv.Itoa(fdbErr.Code), true
	}
	return "", false
}

func getTigrisError(err error) (string, bool) {
	var tigrisErr *api.TigrisError
	if errors.As(err, &tigrisErr) {
		return tigrisErr.Code.String(), true
	}
	return "", false
}

func getTagsForError(err error) map[string]string {
	value, isFdbError := getFdbError(err)
	if isFdbError {
		return map[string]string{
			"error_source": "fdb",
			"error_value":  value,
		}
	}

	value, isTigrisError := getTigrisError(err)
	if isTigrisError {
		return map[string]string{
			"error_source": "tigris_server",
			"error_value":  value,
		}
	}
	// TODO: handle search errors
	return map[string]string{
		"error_source": "unknown",
		"error_value":  "unknown",
	}
}

func getDbTags(dbName string) map[string]string {
	return map[string]string{
		"db": dbName,
	}
}

func getDbCollTags(dbName string, collName string) map[string]string {
	return map[string]string{
		"db":         dbName,
		"collection": collName,
	}
}

func GetDbCollTagsForReq(req interface{}) map[string]string {
	if rc, ok := req.(api.RequestWithDbAndCollection); ok {
		return getDbCollTags(rc.GetDb(), rc.GetCollection())
	}
	if r, ok := req.(api.RequestWithDb); ok {
		return getDbTags(r.GetDb())
	}
	return map[string]string{}
}

func standardizeTags(tags map[string]string, stdKeys []string) map[string]string {
	res := tags
	for _, tagKey := range stdKeys {
		if _, ok := tags[tagKey]; !ok {
			// tag is missing, need to add it
			res[tagKey] = UnknownValue
		} else {
			if res[tagKey] == "" {
				res[tagKey] = UnknownValue
			}
		}
	}
	for k := range res {
		extraTag := true
		// result has an extra tag that should not be there
		for _, stdKey := range stdKeys {
			if stdKey == k {
				extraTag = false
			}
		}
		if extraTag {
			delete(res, k)
		}
	}
	return res
}
