package otel

// NOTE(jm): we have to define this here, because the `plogotlp.ExportRequest` type is actually a hidden type and means
// we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/logs.json#L67

// NOTE(fd): Attributes can be key, and StringValue, IntValue, BoolValue, ArrayValue, KeylistValue, etc.
// this struct is not used for validation of incoming or outgoing data, it is just used in our API docs.
// validation takes place when we unmarsal the json w/ expreq.UnmarshalJSON.

// we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples

type Attribute struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value,omitempty"`
	Value0 struct {
		BoolValue bool `json:"boolValue"`
	} `json:"value,omitempty"`
	Value1 struct {
		IntValue string `json:"intValue"`
	} `json:"value,omitempty"`
	Value2 struct {
		DoubleValue float64 `json:"doubleValue"`
	} `json:"value,omitempty"`
	Value3 struct {
		ArrayValue struct {
			Values []struct {
				StringValue string `json:"stringValue"`
			} `json:"values"`
		} `json:"arrayValue"`
	} `json:"value,omitempty"`
	Value4 struct {
		KvlistValue struct {
			Values []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"values"`
		} `json:"kvlistValue"`
	} `json:"value,omitempty"`
}
type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Scope struct {
	Name                   string      `json:"name,omitempty"`
	Version                string      `json:"version,omitempty"`
	Attributes             []Attribute `json:"attributes,omitempty"`
	DroppedAttributesCount uint32      `json:"droppedAttributesCount,omitempty"`
}

type Body struct {
	StringValue string `json:"stringValue"`
}

type OTLPLogExportRequest struct {
	ResourceLogs []struct {
		Resource  `json:"resource"`
		ScopeLogs []struct {
			SchemaURL  string `json:"schemaUrl,omitempty"`
			Scope      Scope  `json:"scope"`
			LogRecords []struct {
				TimeUnixNano           string      `json:"timeUnixNano"`
				SeverityNumber         int         `json:"severityNumber"`
				SeverityText           string      `json:"severityText"`
				ServiceName            string      `json:"serviceName"`
				Flags                  int         `json:"flags,omitempty"`
				DroppedAttributesCount int         `json:"droppedAttributesCount"`
				TraceID                string      `json:"traceId"`
				SpanID                 string      `json:"spanId"`
				Body                   Body        `json:"body"`
				Attributes             []Attribute `json:"attributes"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	} `json:"resourceLogs"`
}

// NOTE(jm): we have to define this here, because the `pmetricotlp.ExportRequest` type is actually a hidden type and
// means we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples
type OTLPMetricExportRequest struct {
	ResourceMetrics []struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeMetrics []struct {
			Scope   struct{} `json:"scope"`
			Metrics []struct {
				Name string `json:"name"`
				Unit string `json:"unit"`
				Sum  struct {
					DataPoints []struct {
						Attributes []struct {
							Key   string `json:"key"`
							Value struct {
								StringValue string `json:"stringValue"`
							} `json:"value"`
						} `json:"attributes"`
						StartTimeUnixNano string `json:"startTimeUnixNano"`
						TimeUnixNano      string `json:"timeUnixNano"`
						AsInt             string `json:"asInt"`
					} `json:"dataPoints"`
					AggregationTemporality int  `json:"aggregationTemporality"`
					IsMonotonic            bool `json:"isMonotonic"`
				} `json:"sum"`
			} `json:"metrics"`
		} `json:"scopeMetrics"`
	} `json:"resourceMetrics"`
}

// NOTE(jm): we have to define this here, because the `ptraceotlp.ExportRequest` type is actually a hidden type and
// means we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/trace.json
type OTLPTraceExportRequest struct {
	ResourceSpans []struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeSpans []struct {
			Scope struct {
				Name       string `json:"name"`
				Version    string `json:"version"`
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"scope"`
			Spans []struct {
				TraceID           string `json:"traceId"`
				SpanID            string `json:"spanId"`
				ParentSpanID      string `json:"parentSpanId"`
				Name              string `json:"name"`
				StartTimeUnixNano string `json:"startTimeUnixNano"`
				EndTimeUnixNano   string `json:"endTimeUnixNano"`
				Kind              int    `json:"kind"`
				Attributes        []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"spans"`
		} `json:"scopeSpans"`
	} `json:"resourceSpans"`
}
