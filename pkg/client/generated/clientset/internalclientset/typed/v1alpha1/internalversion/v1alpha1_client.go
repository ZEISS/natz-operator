/*
Copyright 2024.

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

// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	http "net/http"

	scheme "github.com/zeiss/natz-operator/pkg/client/generated/clientset/internalclientset/scheme"
	rest "k8s.io/client-go/rest"
)

type NatzInterface interface {
	RESTClient() rest.Interface
	NatsAccountsGetter
	NatsActivationsGetter
	NatsConfigsGetter
	NatsGatewaysGetter
	NatsKeysGetter
	NatsOperatorsGetter
	NatsUsersGetter
}

// NatzClient is used to interact with features provided by the natz.zeiss.com group.
type NatzClient struct {
	restClient rest.Interface
}

func (c *NatzClient) NatsAccounts() NatsAccountInterface {
	return newNatsAccounts(c)
}

func (c *NatzClient) NatsActivations() NatsActivationInterface {
	return newNatsActivations(c)
}

func (c *NatzClient) NatsConfigs() NatsConfigInterface {
	return newNatsConfigs(c)
}

func (c *NatzClient) NatsGateways() NatsGatewayInterface {
	return newNatsGateways(c)
}

func (c *NatzClient) NatsKeys() NatsKeyInterface {
	return newNatsKeys(c)
}

func (c *NatzClient) NatsOperators() NatsOperatorInterface {
	return newNatsOperators(c)
}

func (c *NatzClient) NatsUsers() NatsUserInterface {
	return newNatsUsers(c)
}

// NewForConfig creates a new NatzClient for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*NatzClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new NatzClient for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*NatzClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &NatzClient{client}, nil
}

// NewForConfigOrDie creates a new NatzClient for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *NatzClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new NatzClient for the given RESTClient.
func New(c rest.Interface) *NatzClient {
	return &NatzClient{c}
}

func setConfigDefaults(config *rest.Config) error {
	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	if config.GroupVersion == nil || config.GroupVersion.Group != scheme.Scheme.PrioritizedVersionsForGroup("natz.zeiss.com")[0].Group {
		gv := scheme.Scheme.PrioritizedVersionsForGroup("natz.zeiss.com")[0]
		config.GroupVersion = &gv
	}
	config.NegotiatedSerializer = rest.CodecFactoryForGeneratedClient(scheme.Scheme, scheme.Codecs)

	if config.QPS == 0 {
		config.QPS = 5
	}
	if config.Burst == 0 {
		config.Burst = 10
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *NatzClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
