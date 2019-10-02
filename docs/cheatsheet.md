# OpenTracing Cheat Sheet

For your convenience we have created this cheat sheet to help you with the different steps we will go through during the Workshop.


## Span creation

### Java

Creating a root span

```
Span span = tracer.buildSpan(OPERATION_NAME).start();
```

With a Span Context

```
Span span = tracer.buildSpan(OPERATION_NAME).asChildOf(spanContext).start();
```

From another span created in the current operation.

```
Span cartApiCallSpan = tracer.buildSpan(OPERATION_NAME).asChildOf(someOtherSpan).start();
```

### Golang

Creating a root span

```
span = api.tracer.StartSpan(OPERATION_NAME)
```

With a span context

```
span := api.tracer.StartSpan(OPERATION_NAME, ext.RPCServerOption(spanContext))
```

From another span created in the current operation

```
span := tracer.StartSpan(OPERATION_NAME, opentracing.ChildOf(parentSpan.Context()))
```


## Span context extraction

### Java

```
Map<String, String> requestHeaders = Collections.list(request.getHeaderNames())
                .stream()
                .collect(Collectors.toMap(h -> h, request::getHeader));

SpanContext spanContext = tracer.extract(HTTP_HEADERS, new TextMapAdapter(requestHeaders));
```

### Golang

```
req := c.Request
ctx, err := api.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
```


## Span context injection

### Java

```
Map<String, String> map = new HashMap<>();
tracer.inject(cartApiCallSpan.context(), HTTP_HEADERS, new TextMapAdapter(map));

HttpHeaders headers = new HttpHeaders();
headers.setAll(map);
```

### Golang

```
httpReq, _ := http.NewRequest("GET", url, nil)
api.tracer.Inject(
	span.Context(),
	opentracing.HTTPHeaders,
	opentracing.HTTPHeadersCarrier(httpReq.Header))
```


## Creating application specific tags

### Java

```
span.setTag("az", "1a");
```

### Golang

```
span.SetTag("az", "1a")
```


## Creating semantic convention tags (Ext package)

### Java

```
Tags.SPAN_KIND.set(span, Tags.SPAN_KIND_SERVER);

Tags.HTTP_STATUS.set(span, 200);

Tags.ERROR.set(span, true);
```

### Golang

```
ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)

ext.HTTPStatusCode.Set(span, 200)

ext.Error.Set(span, true)
```


## Adding logs

### Java

```
span.log("some message we wish to log");

span.log(ImmutableMap.of("some key", "some value"));
```

### Golang

```
span.LogKV("some key", "some value")
```


## Finishing span

### Java

```
span.finish();
```

### Golang

```
span.Finish()
```


## Further Links

[Opentracing.io](https://opentracing.io)
