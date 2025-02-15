/*
Copyright 2023 The Kubernetes Authors.

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

package options

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"
	oteltrace "go.opentelemetry.io/otel/trace"

	"k8s.io/apiserver/pkg/admission"
	apiserveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/etcd3"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	auditbuffered "k8s.io/apiserver/plugin/pkg/audit/buffered"
	audittruncate "k8s.io/apiserver/plugin/pkg/audit/truncate"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/metrics"
	netutils "k8s.io/utils/net"

	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

func TestAddFlags(t *testing.T) {
	fs := pflag.NewFlagSet("addflagstest", pflag.PanicOnError)
	s := NewOptions()
	var fss cliflag.NamedFlagSets
	s.AddFlags(&fss)
	for _, f := range fss.FlagSets {
		fs.AddFlagSet(f)
	}

	args := []string{
		"--enable-admission-plugins=AlwaysDeny",
		"--admission-control-config-file=/admission-control-config",
		"--advertise-address=192.168.10.10",
		"--anonymous-auth=false",
		"--audit-log-maxage=11",
		"--audit-log-maxbackup=12",
		"--audit-log-maxsize=13",
		"--audit-log-path=/var/log",
		"--audit-log-mode=blocking",
		"--audit-log-batch-buffer-size=46",
		"--audit-log-batch-max-size=47",
		"--audit-log-batch-max-wait=48s",
		"--audit-log-batch-throttle-enable=true",
		"--audit-log-batch-throttle-qps=49.5",
		"--audit-log-batch-throttle-burst=50",
		"--audit-log-truncate-enabled=true",
		"--audit-log-truncate-max-batch-size=45",
		"--audit-log-truncate-max-event-size=44",
		"--audit-log-version=audit.k8s.io/v1",
		"--audit-policy-file=/policy",
		"--audit-webhook-config-file=/webhook-config",
		"--audit-webhook-mode=blocking",
		"--audit-webhook-batch-buffer-size=42",
		"--audit-webhook-batch-max-size=43",
		"--audit-webhook-batch-max-wait=1s",
		"--audit-webhook-batch-throttle-enable=false",
		"--audit-webhook-batch-throttle-qps=43.5",
		"--audit-webhook-batch-throttle-burst=44",
		"--audit-webhook-truncate-enabled=true",
		"--audit-webhook-truncate-max-batch-size=43",
		"--audit-webhook-truncate-max-event-size=42",
		"--audit-webhook-initial-backoff=2s",
		"--audit-webhook-version=audit.k8s.io/v1",
		"--authentication-token-webhook-cache-ttl=3m",
		"--authentication-token-webhook-config-file=/token-webhook-config",
		"--authorization-mode=AlwaysDeny,RBAC",
		"--authorization-policy-file=/policy",
		"--authorization-webhook-cache-authorized-ttl=3m",
		"--authorization-webhook-cache-unauthorized-ttl=1m",
		"--authorization-webhook-config-file=/webhook-config",
		"--bind-address=192.168.10.20",
		"--client-ca-file=/client-ca",
		"--cors-allowed-origins=10.10.10.100,10.10.10.200",
		"--contention-profiling=true",
		"--egress-selector-config-file=/var/run/kubernetes/egress-selector/connectivity.yaml",
		"--enable-aggregator-routing=true",
		"--enable-priority-and-fairness=false",
		"--enable-logs-handler=false",
		"--etcd-keyfile=/var/run/kubernetes/etcd.key",
		"--etcd-certfile=/var/run/kubernetes/etcdce.crt",
		"--etcd-cafile=/var/run/kubernetes/etcdca.crt",
		"--http2-max-streams-per-connection=42",
		"--tracing-config-file=/var/run/kubernetes/tracing_config.yaml",
		"--proxy-client-cert-file=/var/run/kubernetes/proxy.crt",
		"--proxy-client-key-file=/var/run/kubernetes/proxy.key",
		"--request-timeout=2m",
		"--storage-backend=etcd3",
		"--lease-reuse-duration-seconds=100",
	}
	fs.Parse(args)

	// This is a snapshot of expected options parsed by args.
	expected := &Options{
		GenericServerRunOptions: &apiserveroptions.ServerRunOptions{
			AdvertiseAddress:            netutils.ParseIPSloppy("192.168.10.10"),
			CorsAllowedOriginList:       []string{"10.10.10.100", "10.10.10.200"},
			MaxRequestsInFlight:         400,
			MaxMutatingRequestsInFlight: 200,
			RequestTimeout:              time.Duration(2) * time.Minute,
			MinRequestTimeout:           1800,
			JSONPatchMaxCopyBytes:       int64(3 * 1024 * 1024),
			MaxRequestBodyBytes:         int64(3 * 1024 * 1024),
		},
		Admission: &kubeoptions.AdmissionOptions{
			GenericAdmission: &apiserveroptions.AdmissionOptions{
				RecommendedPluginOrder: s.Admission.GenericAdmission.RecommendedPluginOrder,
				DefaultOffPlugins:      s.Admission.GenericAdmission.DefaultOffPlugins,
				EnablePlugins:          []string{"AlwaysDeny"},
				ConfigFile:             "/admission-control-config",
				Plugins:                s.Admission.GenericAdmission.Plugins,
				Decorators:             s.Admission.GenericAdmission.Decorators,
			},
		},
		Etcd: &apiserveroptions.EtcdOptions{
			StorageConfig: storagebackend.Config{
				Type: "etcd3",
				Transport: storagebackend.TransportConfig{
					ServerList:     nil,
					KeyFile:        "/var/run/kubernetes/etcd.key",
					TrustedCAFile:  "/var/run/kubernetes/etcdca.crt",
					CertFile:       "/var/run/kubernetes/etcdce.crt",
					TracerProvider: oteltrace.NewNoopTracerProvider(),
				},
				Paging:                true,
				Prefix:                "/registry",
				CompactionInterval:    storagebackend.DefaultCompactInterval,
				CountMetricPollPeriod: time.Minute,
				DBMetricPollInterval:  storagebackend.DefaultDBMetricPollInterval,
				HealthcheckTimeout:    storagebackend.DefaultHealthcheckTimeout,
				ReadycheckTimeout:     storagebackend.DefaultReadinessTimeout,
				LeaseManagerConfig: etcd3.LeaseManagerConfig{
					ReuseDurationSeconds: 100,
					MaxObjectCount:       1000,
				},
			},
			DefaultStorageMediaType: "application/vnd.kubernetes.protobuf",
			DeleteCollectionWorkers: 1,
			EnableGarbageCollection: true,
			EnableWatchCache:        true,
			DefaultWatchCacheSize:   100,
		},
		SecureServing: (&apiserveroptions.SecureServingOptions{
			BindAddress: netutils.ParseIPSloppy("192.168.10.20"),
			BindPort:    6443,
			ServerCert: apiserveroptions.GeneratableKeyCert{
				CertDirectory: "/var/run/kubernetes",
				PairName:      "apiserver",
			},
			HTTP2MaxStreamsPerConnection: 42,
			Required:                     true,
		}).WithLoopback(),
		EventTTL: 1 * time.Hour,
		Audit: &apiserveroptions.AuditOptions{
			LogOptions: apiserveroptions.AuditLogOptions{
				Path:       "/var/log",
				MaxAge:     11,
				MaxBackups: 12,
				MaxSize:    13,
				Format:     "json",
				BatchOptions: apiserveroptions.AuditBatchOptions{
					Mode: "blocking",
					BatchConfig: auditbuffered.BatchConfig{
						BufferSize:     46,
						MaxBatchSize:   47,
						MaxBatchWait:   48 * time.Second,
						ThrottleEnable: true,
						ThrottleQPS:    49.5,
						ThrottleBurst:  50,
					},
				},
				TruncateOptions: apiserveroptions.AuditTruncateOptions{
					Enabled: true,
					TruncateConfig: audittruncate.Config{
						MaxBatchSize: 45,
						MaxEventSize: 44,
					},
				},
				GroupVersionString: "audit.k8s.io/v1",
			},
			WebhookOptions: apiserveroptions.AuditWebhookOptions{
				ConfigFile: "/webhook-config",
				BatchOptions: apiserveroptions.AuditBatchOptions{
					Mode: "blocking",
					BatchConfig: auditbuffered.BatchConfig{
						BufferSize:     42,
						MaxBatchSize:   43,
						MaxBatchWait:   1 * time.Second,
						ThrottleEnable: false,
						ThrottleQPS:    43.5,
						ThrottleBurst:  44,
						AsyncDelegate:  true,
					},
				},
				TruncateOptions: apiserveroptions.AuditTruncateOptions{
					Enabled: true,
					TruncateConfig: audittruncate.Config{
						MaxBatchSize: 43,
						MaxEventSize: 42,
					},
				},
				InitialBackoff:     2 * time.Second,
				GroupVersionString: "audit.k8s.io/v1",
			},
			PolicyFile: "/policy",
		},
		Features: &apiserveroptions.FeatureOptions{
			EnableProfiling:           true,
			EnableContentionProfiling: true,
		},
		Authentication: &kubeoptions.BuiltInAuthenticationOptions{
			Anonymous: &kubeoptions.AnonymousAuthenticationOptions{
				Allow: false,
			},
			ClientCert: &apiserveroptions.ClientCertAuthenticationOptions{
				ClientCA: "/client-ca",
			},
			WebHook: &kubeoptions.WebHookAuthenticationOptions{
				CacheTTL:     180000000000,
				ConfigFile:   "/token-webhook-config",
				Version:      "v1beta1",
				RetryBackoff: apiserveroptions.DefaultAuthWebhookRetryBackoff(),
			},
			BootstrapToken: &kubeoptions.BootstrapTokenAuthenticationOptions{},
			OIDC: &kubeoptions.OIDCAuthenticationOptions{
				UsernameClaim: "sub",
				SigningAlgs:   []string{"RS256"},
			},
			RequestHeader: &apiserveroptions.RequestHeaderAuthenticationOptions{},
			ServiceAccounts: &kubeoptions.ServiceAccountAuthenticationOptions{
				Lookup:           true,
				ExtendExpiration: true,
			},
			TokenFile:            &kubeoptions.TokenFileAuthenticationOptions{},
			TokenSuccessCacheTTL: 10 * time.Second,
			TokenFailureCacheTTL: 0,
		},
		Authorization: &kubeoptions.BuiltInAuthorizationOptions{
			Modes:                       []string{"AlwaysDeny", "RBAC"},
			PolicyFile:                  "/policy",
			WebhookConfigFile:           "/webhook-config",
			WebhookCacheAuthorizedTTL:   180000000000,
			WebhookCacheUnauthorizedTTL: 60000000000,
			WebhookVersion:              "v1beta1",
			WebhookRetryBackoff:         apiserveroptions.DefaultAuthWebhookRetryBackoff(),
		},
		APIEnablement: &apiserveroptions.APIEnablementOptions{
			RuntimeConfig: cliflag.ConfigurationMap{},
		},
		EgressSelector: &apiserveroptions.EgressSelectorOptions{
			ConfigFile: "/var/run/kubernetes/egress-selector/connectivity.yaml",
		},
		EnableLogsHandler:       false,
		EnableAggregatorRouting: true,
		ProxyClientKeyFile:      "/var/run/kubernetes/proxy.key",
		ProxyClientCertFile:     "/var/run/kubernetes/proxy.crt",
		Metrics:                 &metrics.Options{},
		Logs:                    logs.NewOptions(),
		Traces: &apiserveroptions.TracingOptions{
			ConfigFile: "/var/run/kubernetes/tracing_config.yaml",
		},
		AggregatorRejectForwardingRedirects: true,
		SystemNamespaces:                    []string{"kube-system", "kube-public", "default"},
	}

	if !reflect.DeepEqual(expected, s) {
		t.Errorf("Got different run options than expected.\nDifference detected on:\n%s", cmp.Diff(expected, s, cmpopts.IgnoreUnexported(admission.Plugins{})))
	}
}
