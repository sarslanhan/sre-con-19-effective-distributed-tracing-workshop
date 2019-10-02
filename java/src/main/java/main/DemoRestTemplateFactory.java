package main;

import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.web.client.ResponseErrorHandler;
import org.springframework.web.client.RestTemplate;

import java.io.IOException;

public class DemoRestTemplateFactory {
    public static RestTemplate createDemoRestTemplate() {
        return new RestTemplateBuilder()
                .errorHandler(new DemoResponseErrorHandler())
                .build();
    }
    
    static class DemoResponseErrorHandler implements ResponseErrorHandler {

        @Override
        public boolean hasError(final ClientHttpResponse response) throws IOException {
            return false;
        }

        @Override
        public void handleError(final ClientHttpResponse response) throws IOException {
            // Don't go breaking my demo!
        }
    }
}
