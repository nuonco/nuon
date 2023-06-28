# Timestamp Data Types and Representation

The timestamps we use in first-party application code flow through the following representation in our stack:

In postgres, the column type is `timestampz` ([docs](https://www.postgresql.org/docs/current/datatype-datetime.html#DATATYPE-DATETIME-INPUT)) which is an abbreviation postgres accepts for the official SQL data type `"timestamp with time zone"`. This is an 8-byte representation in postgres including date, time to 1 microsecond precision, and time zone. All of our times in postgres SHOULD be stored at UTC. (Pete confirmed this to be true as of 2023-05-24).

In golang using gorm, our structs use the go standard library `time.Time` struct.

At the gRPC/protobuf layer, we use the `google.type.DateTime` third party data type ([docs](https://pkg.go.dev/google.golang.org/genproto/googleapis/type/datetime)). This library is published by google [at github.com/googleapis/googleapis](https://github.com/googleapis/googleapis/blob/master/google/type/datetime.proto) and vendored by the buf schema registry at [buf.build/googleapis/googleapis](https://buf.build/googleapis/googleapis). The docs have a comment `// This type is more flexible than some applications may want.` and I suspect that may apply to our specific case. We may want to opt for a simpler type that has either a numeric or string representation at some point. But for the moment, this is what we use.

With our protoc golang code, the go struct for this is "google.golang.org/genproto/googleapis/type/datetime" `datetime.DateTime` and we have a conversion function. This function forces the time into UTC as a precaution in case a local timezone timestamp gets written into postgres somehow by mistake. Thus everything coming back in gRPC replies should be a UTC timestamp.

In the API Gateway, we use the buf-generated javascript implementation of `google.type.DateTime` which has the `google_type_datetime_pb.DateTime.toObject` function. We model this object in typescript with a custom type `TDateTimeObject` for static typing purposes. We convert this plain object into a [luxon DateTime](https://moment.github.io/luxon/api-docs/index.html#datetimefromobject) and then to a string with `.toISO()`. All ISO timestamp strings coming out of the API gateway currently should be in UTC and in this syntax: `2023-05-17T21:51:15.000Z`.
