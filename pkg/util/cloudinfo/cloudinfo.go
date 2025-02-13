// Copyright 2019 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package cloudinfo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

const (
	aws   = "Amazon Web Services"
	gcp   = "Google Cloud Platform"
	azure = "Microsoft Azure"
)

// parseAWSInstanceMetadata uses the structure described
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
// If we encounter JSON we cannot marhsal into this structure, we
// assume we're not running on AWS.
func parseAWSInstanceMetadata(body []byte) (bool, string, string) {
	instanceMetadata := struct {
		InstanceClass string `json:"instanceType"`
	}{}

	success := true
	if err := json.Unmarshal(body, &instanceMetadata); err != nil {
		success = false
	}

	return success, aws, instanceMetadata.InstanceClass
}

// parseGCPInstanceMetadata relies on the structure indicated at
// https://cloud.google.com/compute/docs/storing-retrieving-metadata
// If we encounter a string that doesn't match our format, we  assume
// we're not running on GCP.
func parseGCPInstanceMetadata(body []byte) (bool, string, string) {
	bodyStr := string(body)

	// The structure of the API's response can be found at
	// https://cloud.google.com/compute/docs/storing-retrieving-metadata;
	// look for machine-type
	instanceClassRE := regexp.MustCompile(`machineTypes\/(.+)$`)

	instanceClass := instanceClassRE.FindStringSubmatch(bodyStr)

	// Regex should only have 2 values: matched string and
	// capture group containing the machineTypes value.
	if len(instanceClass) != 2 {
		return false, "", ""
	}

	return true, gcp, instanceClass[1]
}

// parseAzureInstanceMetadata uses the structure described
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service
// If we encounter JSON we cannot marhsal into this structure, we
// assume we're not running on Azure.
func parseAzureInstanceMetadata(body []byte) (bool, string, string) {
	instanceMetadata := struct {
		ComputeEnv struct {
			InstanceClass string `json:"vmSize"`
		} `json:"compute"`
	}{}

	success := true
	if err := json.Unmarshal(body, &instanceMetadata); err != nil {
		success = false
	}

	return success, azure, instanceMetadata.ComputeEnv.InstanceClass
}

type metadataReqHeader struct {
	key   string
	value string
}

func getInstanceMetadata(url string, headers []metadataReqHeader) ([]byte, error) {
	client := http.Client{
		Timeout: 500 * time.Millisecond,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for _, header := range headers {
		req.Header.Set(header.key, header.value)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)

}

// GetProviderInfo returns the node's instance provider (e.g. AWS) and
// the name given to its instance class (e.g. m5a.large).
func GetProviderInfo() (string, string) {

	// providerInstanceMetadataDetails provides all necessary details
	// to make http.Get() request to cloud provider metadata endpoint
	// and get a response as a slice of bytes.
	providerInstanceMetadataDetails := []struct {
		url     string
		headers []metadataReqHeader
		parse   func([]byte) (bool, string, string)
	}{
		// AWS reference https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
		{
			url:   "http://instance-data.ec2.internal/latest/dynamic/instance-identity/document",
			parse: parseAWSInstanceMetadata,
		},
		// GCP reference https://cloud.google.com/compute/docs/storing-retrieving-metadata
		{
			url: "http://metadata.google.internal/computeMetadata/v1/instance/machine-type",
			headers: []metadataReqHeader{{
				"Metadata-Flavor", "Google",
			}},
			parse: parseGCPInstanceMetadata,
		},
		// Azure reference https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service
		{
			url: "http://169.254.169.254/metadata/instance?api-version=2018-10-01",
			headers: []metadataReqHeader{{
				"Metadata", "true",
			}},
			parse: parseAzureInstanceMetadata,
		},
	}

	var success bool
	var providerName, instanceClass string

	for _, p := range providerInstanceMetadataDetails {
		body, err := getInstanceMetadata(p.url, p.headers)

		if err != nil {
			continue
		}
		success, providerName, instanceClass = p.parse(body)
		if success {
			return providerName, instanceClass
		}
	}

	return "", ""
}
