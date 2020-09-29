# Titan

[![Build Status](https://gitlab.com/silenteer-oss/titan/badges/master/pipeline.svg)](https://gitlab.com/silenteer-oss/titan/badges/master/pipeline.svg)

Titan is an RPC framework like [gRPC](https://grpc.io/), but it uses
[NATS](https://nats.io/) as transport layer.

It can use in combination with [goff](https://gitlab.com/silenteer-oss/goff)  to generate a Go client and server from the same .proto file that you'd
use to generate gRPC clients and servers. The server is generated as a NATS
[MsgHandler](https://godoc.org/github.com/nats-io/nats.go#MsgHandler).

## The key features?
- **Lightweight**: NATS is small less than ~10MB
- **Service Discovery**: Built-in by NATS subject.
- **Load Balancing**: Built-in by NATS subject and queue group.
- **High performant**: NATS can handle millions of requests per second.
- **Scalability**: NATS clustering 
- **Request & Reply**: 
- **Publish & Subscribe**: 
- **Tracing**: Integrated with [Jaeger](https://github.com/jaegertracing/jaeger) (under construction).
- **Monitoring**: NATS monitoring dashboard.
- **Payload validation**: https://github.com/go-playground/validator
- **Serialization**: Using json
- **Metadata**: contextual data is transfer across services.
- **Authentication**: Role base checking.
