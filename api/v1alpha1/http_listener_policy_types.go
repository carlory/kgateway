package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// +kubebuilder:rbac:groups=gateway.kgateway.dev,resources=httplistenerpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.kgateway.dev,resources=httplistenerpolicies/status,verbs=get;update;patch

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:metadata:labels={app=kgateway,app.kubernetes.io/name=kgateway}
// +kubebuilder:resource:categories=kgateway
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="gateway.networking.k8s.io/policy=Direct"
type HTTPListenerPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HTTPListenerPolicySpec `json:"spec,omitempty"`

	Status gwv1alpha2.PolicyStatus `json:"status,omitempty"`
	// TODO: embed this into a typed Status field when
	// https://github.com/kubernetes/kubernetes/issues/131533 is resolved
}

// +kubebuilder:object:root=true
type HTTPListenerPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPListenerPolicy `json:"items"`
}

type HTTPListenerPolicySpec struct {
	// TargetRefs specifies the target resources by reference to attach the policy to.
	// +optional
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	TargetRefs []LocalPolicyTargetReference `json:"targetRefs,omitempty"`

	// TargetSelectors specifies the target selectors to select resources to attach the policy to.
	// +optional
	TargetSelectors []LocalPolicyTargetSelector `json:"targetSelectors,omitempty"`

	// AccessLoggingConfig contains various settings for Envoy's access logging service.
	// See here for more information: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto
	// +kubebuilder:validation:Items={type=object}
	AccessLog []AccessLog `json:"accessLog,omitempty"`

	// UpgradeConfig contains configuration for HTTP upgrades like WebSocket.
	// See here for more information: https://www.envoyproxy.io/docs/envoy/v1.34.1/intro/arch_overview/http/upgrades.html
	UpgradeConfig *UpgradeConfig `json:"upgradeConfig,omitempty"`

	// Tracing contains various settings for Envoy's OpenTelemetry tracer.
	// See here for more information: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/trace/v3/opentelemetry.proto.html
	Tracing *Tracing `json:"tracing,omitempty"`
}

// AccessLog represents the top-level access log configuration.
type AccessLog struct {
	// Output access logs to local file
	FileSink *FileSink `json:"fileSink,omitempty"`

	// Send access logs to gRPC service
	GrpcService *AccessLogGrpcService `json:"grpcService,omitempty"`

	// Send access logs to an OTel collector
	OpenTelemetry *OpenTelemetryAccessLogService `json:"openTelemetry,omitempty"`

	// Filter access logs configuration
	Filter *AccessLogFilter `json:"filter,omitempty"`
}

// FileSink represents the file sink configuration for access logs.
// +kubebuilder:validation:XValidation:message="only one of 'StringFormat' or 'JsonFormat' may be set",rule="(has(self.stringFormat) && !has(self.jsonFormat)) || (!has(self.stringFormat) && has(self.jsonFormat))"
type FileSink struct {
	// the file path to which the file access logging service will sink
	// +required
	Path string `json:"path"`
	// the format string by which envoy will format the log lines
	// https://www.envoyproxy.io/docs/envoy/v1.33.0/configuration/observability/access_log/usage#format-strings
	StringFormat string `json:"stringFormat,omitempty"`
	// the format object by which to envoy will emit the logs in a structured way.
	// https://www.envoyproxy.io/docs/envoy/v1.33.0/configuration/observability/access_log/usage#format-dictionaries
	JsonFormat *runtime.RawExtension `json:"jsonFormat,omitempty"`
}

// AccessLogGrpcService represents the gRPC service configuration for access logs.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#envoy-v3-api-msg-extensions-access-loggers-grpc-v3-httpgrpcaccesslogconfig
type AccessLogGrpcService struct {
	*CommonAccessLogGrpcService `json:",inline"`

	// Additional request headers to log in the access log
	AdditionalRequestHeadersToLog []string `json:"additionalRequestHeadersToLog,omitempty"`

	// Additional response headers to log in the access log
	AdditionalResponseHeadersToLog []string `json:"additionalResponseHeadersToLog,omitempty"`

	// Additional response trailers to log in the access log
	AdditionalResponseTrailersToLog []string `json:"additionalResponseTrailersToLog,omitempty"`
}

// Common configuration for gRPC access logs.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/grpc/v3/als.proto#envoy-v3-api-msg-extensions-access-loggers-grpc-v3-commongrpcaccesslogconfig
type CommonAccessLogGrpcService struct {
	*CommonGrpcService `json:",inline"`

	// name of log stream
	// +kubebuilder:validation:Required
	LogName string `json:"logName"`
}

// Common gRPC service configuration created by setting `envoy_grpc“ as the gRPC client
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/grpc_service.proto#envoy-v3-api-msg-config-core-v3-grpcservice
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/grpc_service.proto#envoy-v3-api-msg-config-core-v3-grpcservice-envoygrpc
type CommonGrpcService struct {
	// The backend gRPC service. Can be any type of supported backend (Kubernetes Service, kgateway Backend, etc..)
	// +kubebuilder:validation:Required
	BackendRef *gwv1.BackendRef `json:"backendRef"`

	// The :authority header in the grpc request. If this field is not set, the authority header value will be cluster_name.
	// Note that this authority does not override the SNI. The SNI is provided by the transport socket of the cluster.
	// +kubebuilder:validation:Optional
	Authority *string `json:"authority,omitempty"`

	// Maximum gRPC message size that is allowed to be received. If a message over this limit is received, the gRPC stream is terminated with the RESOURCE_EXHAUSTED error.
	// Defaults to 0, which means unlimited.
	// +kubebuilder:validation:Optional
	MaxReceiveMessageLength *uint32 `json:"maxReceiveMessageLength,omitempty"`

	// This provides gRPC client level control over envoy generated headers. If false, the header will be sent but it can be overridden by per stream option. If true, the header will be removed and can not be overridden by per stream option. Default to false.
	// +kubebuilder:validation:Optional
	SkipEnvoyHeaders *bool `json:"skipEnvoyHeaders,omitempty"`

	// The timeout for the gRPC request. This is the timeout for a specific request
	// +kubebuilder:validation:Optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// Additional metadata to include in streams initiated to the GrpcService.
	// This can be used for scenarios in which additional ad hoc authorization headers (e.g. x-foo-bar: baz-key) are to be injected
	// +kubebuilder:validation:Optional
	InitialMetadata []HeaderValue `json:"initialMetadata,omitempty"`

	// Indicates the retry policy for re-establishing the gRPC stream.
	// If max interval is not provided, it will be set to ten times the provided base interval
	// +kubebuilder:validation:Optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

// Header name/value pair.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-headervalue
type HeaderValue struct {
	// Header name.
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`

	// Header value.
	// +kubebuilder:validation:Optional
	Value string `json:"value,omitempty"`
}

// Specifies the retry policy of remote data source when fetching fails.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-retrypolicy
type RetryPolicy struct {
	// Specifies parameters that control retry backoff strategy.
	// the default base interval is 1000 milliseconds and the default maximum interval is 10 times the base interval.
	// +kubebuilder:validation:Optional
	RetryBackOff *BackoffStrategy `json:"retryBackOff,omitempty"`

	// Specifies the allowed number of retries. Defaults to 1.
	// +kubebuilder:validation:Optional
	NumRetries *uint32 `json:"numRetries,omitempty"`
}

// Configuration defining a jittered exponential back off strategy.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/backoff.proto#envoy-v3-api-msg-config-core-v3-backoffstrategy
type BackoffStrategy struct {
	// The base interval to be used for the next back off computation. It should be greater than zero and less than or equal to max_interval.
	// +kubebuilder:validation:Required
	BaseInterval metav1.Duration `json:"baseInterval,omitempty"`

	// Specifies the maximum interval between retries. This parameter is optional, but must be greater than or equal to the base_interval if set. The default is 10 times the base_interval.
	// +kubebuilder:validation:Optional
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
}

// OpenTelemetryAccessLogService represents the OTel configuration for access logs.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/access_loggers/open_telemetry/v3/logs_service.proto
type OpenTelemetryAccessLogService struct {
	// Send access logs to gRPC service
	// +kubebuilder:validation:Required
	GrpcService *CommonAccessLogGrpcService `json:"grpcService,omitempty"`

	// OpenTelemetry LogResource fields, following Envoy access logging formatting.
	// +kubebuilder:validation:Optional
	Body *string `json:"body,omitempty"`

	// If specified, Envoy will not generate built-in resource labels like log_name, zone_name, cluster_name, node_name.
	// +kubebuilder:validation:Optional
	DisableBuiltinLabels *bool `json:"disableBuiltinLabels,omitempty"`

	// Additional attributes that describe the specific event occurrence.
	// +kubebuilder:validation:Optional
	Attributes *KeyAnyValueList `json:"attributes,omitempty"`
}

// A list of key-value pair that is used to store Span attributes, Link attributes, etc.
type KeyAnyValueList struct {
	// A collection of key/value pairs of key-value pairs.
	// +kubebuilder:validation:Required
	Values []KeyAnyValue `json:"values,omitempty"`
}

// KeyValue is a key-value pair that is used to store Span attributes, Link attributes, etc.
type KeyAnyValue struct {
	// Attribute keys must be unique
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`
	// Value may contain a primitive value such as a string or integer or it may contain an arbitrary nested object containing arrays, key-value lists and primitives.
	// +kubebuilder:validation:Required
	Value AnyValue `json:"value,omitempty"`
}

// AnyValue is used to represent any type of attribute value. AnyValue may contain a primitive value such as a string or integer or it may contain an arbitrary nested object containing arrays, key-value lists and primitives.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AnyValue struct {
	BoolValue *bool `json:"boolValue,omitempty"`
	// +kubebuilder:validation:Format=numerical
	DoubleValue *string          `json:"doubleValue,omitempty"`
	IntValue    *int64           `json:"intValue,omitempty"`
	StringValue *string          `json:"stringValue,omitempty"`
	BytesValue  []byte           `json:"bytesValue,omitempty"`
	ArrayValue  []AnyValue       `json:"arrayValue,omitempty"`
	KvListValue *KeyAnyValueList `json:"kvListValue,omitempty"`
}

// AccessLogFilter represents the top-level filter structure.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-accesslogfilter
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AccessLogFilter struct {
	*FilterType `json:",inline"` // embedded to allow for validation
	// Performs a logical "and" operation on the result of each individual filter.
	// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-andfilter
	// +kubebuilder:validation:MinItems=2
	AndFilter []FilterType `json:"andFilter,omitempty"`
	// Performs a logical "or" operation on the result of each individual filter.
	// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-orfilter
	// +kubebuilder:validation:MinItems=2
	OrFilter []FilterType `json:"orFilter,omitempty"`
}

// FilterType represents the type of filter to apply (only one of these should be set).
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#envoy-v3-api-msg-config-accesslog-v3-accesslogfilter
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type FilterType struct {
	StatusCodeFilter *StatusCodeFilter `json:"statusCodeFilter,omitempty"`
	DurationFilter   *DurationFilter   `json:"durationFilter,omitempty"`
	// Filters for requests that are not health check requests.
	// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-nothealthcheckfilter
	NotHealthCheckFilter bool `json:"notHealthCheckFilter,omitempty"`
	// Filters for requests that are traceable.
	// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-traceablefilter
	TraceableFilter    bool                `json:"traceableFilter,omitempty"`
	HeaderFilter       *HeaderFilter       `json:"headerFilter,omitempty"`
	ResponseFlagFilter *ResponseFlagFilter `json:"responseFlagFilter,omitempty"`
	GrpcStatusFilter   *GrpcStatusFilter   `json:"grpcStatusFilter,omitempty"`
	CELFilter          *CELFilter          `json:"celFilter,omitempty"`
}

// ComparisonFilter represents a filter based on a comparison.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-comparisonfilter
type ComparisonFilter struct {
	// +required
	Op Op `json:"op,omitempty"`

	// Value to compare against.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	Value uint32 `json:"value,omitempty"`
}

// Op represents comparison operators.
// +kubebuilder:validation:Enum=EQ;GE;LE
type Op string

const (
	EQ Op = "EQ" // Equal
	GE Op = "GQ" // Greater or equal
	LE Op = "LE" // Less or equal
)

// StatusCodeFilter filters based on HTTP status code.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#envoy-v3-api-msg-config-accesslog-v3-statuscodefilter
type StatusCodeFilter ComparisonFilter

// DurationFilter filters based on request duration.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-durationfilter
type DurationFilter ComparisonFilter

// DenominatorType defines the fraction percentages support several fixed denominator values.
// +kubebuilder:validation:enum=HUNDRED,TEN_THOUSAND,MILLION
type DenominatorType string

const (
	// 100.
	//
	// **Example**: 1/100 = 1%.
	HUNDRED DenominatorType = "HUNDRED"
	// 10,000.
	//
	// **Example**: 1/10000 = 0.01%.
	TEN_THOUSAND DenominatorType = "TEN_THOUSAND"
	// 1,000,000.
	//
	// **Example**: 1/1000000 = 0.0001%.
	MILLION DenominatorType = "MILLION"
)

// HeaderFilter filters requests based on headers.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-headerfilter
type HeaderFilter struct {
	// +required
	Header gwv1.HTTPHeaderMatch `json:"header"`
}

// ResponseFlagFilter filters based on response flags.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#config-accesslog-v3-responseflagfilter
type ResponseFlagFilter struct {
	// +kubebuilder:validation:MinItems=1
	Flags []string `json:"flags"`
}

// CELFilter filters requests based on Common Expression Language (CEL).
type CELFilter struct {
	// The CEL expressions to evaluate. AccessLogs are only emitted when the CEL expressions evaluates to true.
	// see: https://www.envoyproxy.io/docs/envoy/v1.33.0/xds/type/v3/cel.proto.html#common-expression-language-cel-proto
	Match string `json:"match"`
}

// GrpcStatusFilter filters gRPC requests based on their response status.
// Based on: https://www.envoyproxy.io/docs/envoy/v1.33.0/api-v3/config/accesslog/v3/accesslog.proto#enum-config-accesslog-v3-grpcstatusfilter-status
type GrpcStatusFilter struct {
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Items={type=object}
	Statuses []GrpcStatus `json:"statuses,omitempty"`
	Exclude  bool         `json:"exclude,omitempty"`
}

// Tracing represents the top-level Envoy's tracer.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-tracing
type Tracing struct {
	// Provider defines the upstream to which envoy sends traces
	// +kubebuilder:validation:Required
	Provider *TracingProvider `json:"provider"`

	// Target percentage of requests managed by this HTTP connection manager that will be force traced if the x-client-trace-id header is set. Defaults to 100%
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	ClientSampling *uint32 `json:"clientSampling,omitempty"`

	// Target percentage of requests managed by this HTTP connection manager that will be randomly selected for trace generation, if not requested by the client or not forced. Defaults to 100%
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	RandomSampling *uint32 `json:"randomSampling,omitempty"`

	// Target percentage of requests managed by this HTTP connection manager that will be traced after all other sampling checks have been applied (client-directed, force tracing, random sampling). Defaults to 100%
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	OverallSampling *uint32 `json:"overallSampling,omitempty"`

	// Whether to annotate spans with additional data. If true, spans will include logs for stream events. Defaults to false
	// +kubebuilder:validation:Optional
	Verbose *bool `json:"verbose,omitempty"`

	// Maximum length of the request path to extract and include in the HttpUrl tag. Used to truncate lengthy request paths to meet the needs of a tracing backend. Default: 256
	// +kubebuilder:validation:Optional
	MaxPathTagLength *uint32 `json:"maxPathTagLength,omitempty"`

	// A list of custom tags with unique tag name to create tags for the active span.
	// +kubebuilder:validation:Optional
	CustomTags []CustomTag `json:"customTags,omitempty"`

	// Create separate tracing span for each upstream request if true. Defaults to false
	// Link to envoy docs for more info
	// +kubebuilder:validation:Optional
	SpawnUpstreamSpan *bool `json:"spawnUpstreamSpan,omitempty"`
}

// Describes custom tags for the active span.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto#envoy-v3-api-msg-type-tracing-v3-customtag
// +kubebuilder:validation:MaxProperties=2
// +kubebuilder:validation:MinProperties=1
type CustomTag struct {
	// Used to populate the tag name
	// +kubebuilder:validation:Required
	Tag string `json:"tag,omitempty"`

	// A literal custom tag.
	// +kubebuilder:validation:Optional
	Literal *CustomTagLiteral `json:"literal,omitempty"`

	// An environment custom tag.
	// +kubebuilder:validation:Optional
	Environment *CustomTagEnvironment `json:"environment,omitempty"`

	// A request header custom tag.
	// +kubebuilder:validation:Optional
	RequestHeader *CustomTagHeader `json:"RequestHeader,omitempty"`

	// A custom tag to obtain tag value from the metadata.
	// +kubebuilder:validation:Optional
	Metadata *CustomTagMetadata `json:"metadata,omitempty"`
}

// Literal type custom tag with static value for the tag value.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto#type-tracing-v3-customtag-literal
type CustomTagLiteral struct {
	// Static literal value to populate the tag value.
	// +kubebuilder:validation:Required
	Value string `json:"value,omitempty"`
}

// Environment type custom tag with environment name and default value.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto#type-tracing-v3-customtag-environment
type CustomTagEnvironment struct {
	// Environment variable name to obtain the value to populate the tag value.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// When the environment variable is not found, the tag value will be populated with this default value if specified,
	// otherwise no tag will be populated.
	// +kubebuilder:validation:Optional
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// Header type custom tag with header name and default value.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto#type-tracing-v3-customtag-header
type CustomTagHeader struct {
	// Header name to obtain the value to populate the tag value.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// When the header does not exist, the tag value will be populated with this default value if specified,
	// otherwise no tag will be populated.
	// +kubebuilder:validation:Optional
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// Metadata type custom tag using MetadataKey to retrieve the protobuf value from Metadata, and populate the tag value with the canonical JSON representation of it.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/tracing/v3/custom_tag.proto#type-tracing-v3-customtag-metadata
type CustomTagMetadata struct {
	// Specify what kind of metadata to obtain tag value from
	// +kubebuilder:validation:Enum=Request;Route;Cluster;Host
	Kind MetadataKind `json:"kind,omitempty"`

	// Metadata key to define the path to retrieve the tag value.
	// +kubebuilder:validation:Required
	MetadataKey *MetadataKey `json:"metadataKey,omitempty"`

	// When no valid metadata is found, the tag value would be populated with this default value if specified, otherwise no tag would be populated.
	// +kubebuilder:validation:Optional
	DefaultValue *string `json:"defaultValue,omitempty"`
}

// Describes different types of metadata sources.
// Ref: https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/metadata/v3/metadata.proto#envoy-v3-api-msg-type-metadata-v3-metadatakind-request
type MetadataKind string

const (
	// Request kind of metadata.
	MetadataKindRequest MetadataKind = "Request"
	// Route kind of metadata.
	MetadataKindRoute MetadataKind = "Route"
	// Cluster kind of metadata.
	MetadataKindCluster MetadataKind = "Cluster"
	// Host kind of metadata.
	MetadataKindHost MetadataKind = "Host"
)

// MetadataKey provides a way to retrieve values from Metadata using a key and a path.
type MetadataKey struct {
	// The key name of the Metadata from which to retrieve the Struct
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`

	// The path used to retrieve a specific Value from the Struct. This can be either a prefix or a full path,
	// depending on the use case
	// +kubebuilder:validation:Required
	Path []MetadataPathSegment `json:"path,omitempty"`
}

// Specifies a segment in a path for retrieving values from Metadata.
type MetadataPathSegment struct {
	// The key used to retrieve the value in the struct
	// +kubebuilder:validation:Required
	Key string `json:"key,omitempty"`
}

// TracingProvider defines the list of providers for tracing
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type TracingProvider struct {
	// Tracing contains various settings for Envoy's OTel tracer.
	// +kubebuilder:validation:Required
	OpenTelemetry *OpenTelemetryTracingConfig `json:"openTelemetry,omitempty"`
}

// OpenTelemetryTracingConfig represents the top-level Envoy's OpenTelemetry tracer.
// See here for more information: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/trace/v3/opentelemetry.proto.html
type OpenTelemetryTracingConfig struct {
	// Send traces to the gRPC service
	// +kubebuilder:validation:Required
	GrpcService *CommonGrpcService `json:"grpcService,omitempty"`

	// The name for the service. This will be populated in the ResourceSpan Resource attributes
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName,omitempty"`

	// An ordered list of resource detectors. Currently supported values are `EnvironmentResourceDetector`
	// +kubebuilder:validation:Optional
	ResourceDetectors []ResourceDetector `json:"resourceDetectors,omitempty"`

	// Specifies the sampler to be used by the OpenTelemetry tracer. This field can be left empty. In this case, the default Envoy sampling decision is used.
	// Currently supported values are `AlwaysOn`
	// +kubebuilder:validation:Optional
	Sampler *Sampler `json:"sampler,omitempty"`
}

// ResourceDetector defines the list of supported ResourceDetectors
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type ResourceDetector struct {
	EnvironmentResourceDetector *EnvironmentResourceDetectorConfig `json:"environmentResourceDetector,omitempty"`
}

// EnvironmentResourceDetectorConfig specified the EnvironmentResourceDetector
type EnvironmentResourceDetectorConfig struct{}

// Sampler defines the list of supported Samplers
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type Sampler struct {
	AlwaysOn *AlwaysOnConfig `json:"alwaysOnConfig,omitempty"`
}

// AlwaysOnConfig specified the AlwaysOn samplerc
type AlwaysOnConfig struct{}

// GrpcStatus represents possible gRPC statuses.
// +kubebuilder:validation:Enum=OK;CANCELED;UNKNOWN;INVALID_ARGUMENT;DEADLINE_EXCEEDED;NOT_FOUND;ALREADY_EXISTS;PERMISSION_DENIED;RESOURCE_EXHAUSTED;FAILED_PRECONDITION;ABORTED;OUT_OF_RANGE;UNIMPLEMENTED;INTERNAL;UNAVAILABLE;DATA_LOSS;UNAUTHENTICATED
type GrpcStatus string

const (
	OK                  GrpcStatus = "OK"
	CANCELED            GrpcStatus = "CANCELED"
	UNKNOWN             GrpcStatus = "UNKNOWN"
	INVALID_ARGUMENT    GrpcStatus = "INVALID_ARGUMENT"
	DEADLINE_EXCEEDED   GrpcStatus = "DEADLINE_EXCEEDED"
	NOT_FOUND           GrpcStatus = "NOT_FOUND"
	ALREADY_EXISTS      GrpcStatus = "ALREADY_EXISTS"
	PERMISSION_DENIED   GrpcStatus = "PERMISSION_DENIED"
	RESOURCE_EXHAUSTED  GrpcStatus = "RESOURCE_EXHAUSTED"
	FAILED_PRECONDITION GrpcStatus = "FAILED_PRECONDITION"
	ABORTED             GrpcStatus = "ABORTED"
	OUT_OF_RANGE        GrpcStatus = "OUT_OF_RANGE"
	UNIMPLEMENTED       GrpcStatus = "UNIMPLEMENTED"
	INTERNAL            GrpcStatus = "INTERNAL"
	UNAVAILABLE         GrpcStatus = "UNAVAILABLE"
	DATA_LOSS           GrpcStatus = "DATA_LOSS"
	UNAUTHENTICATED     GrpcStatus = "UNAUTHENTICATED"
)

// UpgradeConfig represents configuration for HTTP upgrades.
type UpgradeConfig struct {
	// List of upgrade types to enable (e.g. "websocket", "CONNECT", etc.)
	// +kubebuilder:validation:MinItems=1
	EnabledUpgrades []string `json:"enabledUpgrades,omitempty"`
}
