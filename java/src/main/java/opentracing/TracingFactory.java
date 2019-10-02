package opentracing;

import io.jaegertracing.Configuration;
import io.jaegertracing.internal.metrics.NoopMetricsFactory;
import io.jaegertracing.internal.samplers.ConstSampler;
import io.opentracing.Tracer;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Controller;

@Controller
public class TracingFactory {
    @Value("${endpoints.jaeger}")
    private String collectorEndpoint;

    public Tracer createTracer(String serviceName) {

        Configuration configuration = new Configuration(serviceName)
                .withReporter(new Configuration.ReporterConfiguration()
                        .withLogSpans(true)
                        .withSender(new Configuration.SenderConfiguration().withEndpoint(collectorEndpoint))
                .withLogSpans(true))
                .withMetricsFactory(new NoopMetricsFactory())
                .withSampler(new Configuration.SamplerConfiguration()
                        .withType(ConstSampler.TYPE)
                        .withParam(1));
       
        return configuration.getTracerBuilder().build();
    }
}
