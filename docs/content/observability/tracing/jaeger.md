# Jaeger

To enable the Jaeger:

```toml tab="File"
[tracing]
  [tracing.jaeger]
```

```bash tab="CLI"
--tracing
--tracing.jaeger
```

!!! warning
    Traefik is only able to send data over the compact thrift protocol to the [Jaeger agent](https://www.jaegertracing.io/docs/deployment/#agent).

#### `samplingServerURL`

_Required, Default="http://localhost:5778/sampling"_

Sampling Server URL is the address of jaeger-agent's HTTP sampling server.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingServerURL = "http://localhost:5778/sampling"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingServerURL="http://localhost:5778/sampling"
```

#### `samplingType`

_Required, Default="const"_

Sampling Type specifies the type of the sampler: `const`, `probabilistic`, `rateLimiting`.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingType = "const"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingType="const"
```

#### `samplingParam`

_Required, Default=1.0_

Sampling Param is a value passed to the sampler.

Valid values for Param field are:

- for `const` sampler, 0 or 1 for always false/true respectively
- for `probabilistic` sampler, a probability between 0 and 1
- for `rateLimiting` sampler, the number of spans per second

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingParam = 1.0
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingParam="1.0"
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:6831"_

Local Agent Host Port instructs reporter to send spans to jaeger-agent at this address.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    localAgentHostPort = "127.0.0.1:6831"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.localAgentHostPort="127.0.0.1:6831"
```

#### `gen128Bit`

_Optional, Default=false_

Generate 128-bit trace IDs, compatible with OpenCensus.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    gen128Bit = true
```

```bash tab="CLI"
--tracing
--tracing.jaeger.gen128Bit
```

#### `propagation`

_Required, Default="jaeger"_

Set the propagation header type.
This can be either:

- `jaeger`, jaeger's default trace header.
- `b3`, compatible with OpenZipkin

```toml tab="File"
[tracing]
  [tracing.jaeger]
    propagation = "jaeger"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.propagation="jaeger"
```

#### `traceContextHeaderName`

_Required, Default="uber-trace-id"_

Trace Context Header Name is the http header name used to propagate tracing context.
This must be in lower-case to avoid mismatches when decoding incoming headers.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    traceContextHeaderName = "uber-trace-id"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.traceContextHeaderName="uber-trace-id"
```
