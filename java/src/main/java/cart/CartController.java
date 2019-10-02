package cart;


import com.fasterxml.jackson.databind.JsonNode;
import com.google.common.collect.ImmutableMap;
import fault.FaultInjectionManager;
import io.opentracing.Span;
import io.opentracing.SpanContext;
import io.opentracing.Tracer;
import io.opentracing.propagation.Format;
import io.opentracing.propagation.TextMapAdapter;
import io.opentracing.tag.Tags;
import main.DemoRestTemplateFactory;
import opentracing.TracingFactory;
import opentracing.TracingTags;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.ui.ModelMap;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Collectors;

import static io.opentracing.propagation.Format.Builtin.HTTP_HEADERS;

@Controller
public class CartController {

    private static final Logger LOG = LoggerFactory.getLogger(CartController.class);

    private static final String COMPONENT_NAME = "cart-api";
    private static final String OPERATION_NAME = "add_to_cart";
    private static final String BUSINESS_LOGIC_OPERATION_NAME = "do_complex_business_logic";
    private static final String RECORD_STORAGE_OPERATION_NAME = "store_records";
    private static final String CHECK_STOCK_OPERATION_NAME = "check_stock";
    private static final String REMOTE_STORAGE_FAILURE = "failed to add record to remote storage";
    private static final String STOCK_API_ERROR = "error calling Stock API";
    private static final String SKU_OUT_OF_STOCK = "sku out of stock";

    @Value("${ot.az}")
    private int az;

    @Value("${ot.instance_id}")
    private String instanceId;

    @Value("${endpoints.stock}")
    private String stockApiEndpoint;
    
    private final Tracer tracer;

    private final FaultInjectionManager faultInjectionManager;

    private final RestTemplate restTemplate = DemoRestTemplateFactory.createDemoRestTemplate();

    @Autowired
    public CartController(final FaultInjectionManager faultInjectionManager, final TracingFactory tracingFactory) {
        this.faultInjectionManager = faultInjectionManager;
        this.tracer = tracingFactory.createTracer(COMPONENT_NAME);
    }

    @RequestMapping(value = "/cart/{sku}", method = RequestMethod.POST)
    @ResponseBody
    public Boolean addToCart(@PathVariable(value = "sku") String sku, ModelMap model, HttpServletRequest request) throws InterruptedException {

         complexBusinessLogic();

         boolean success = storeRecordsInRemoteStorage();
         if (!success) {
             return false;
         }

         boolean hasStock = checkStock(sku);
         if (!hasStock) {
             return false;
         } else {
             return true;
         }
            
    }
    
    private boolean checkStock(final String sku) throws InterruptedException {
        String url = stockApiEndpoint + sku;
        ResponseEntity<JsonNode> stockResponse = restTemplate.exchange(url,
                HttpMethod.GET,
                null,
                JsonNode.class);
        Thread.sleep(50);

        if (stockResponse.getStatusCode().isError()) {
            return false;
        }

        JsonNode jsonResponse = stockResponse.getBody();
        if (!jsonResponse.has(sku) || jsonResponse.get(sku).asInt() < 1) {
            return false;
        } else {
            return true;
        }
    }

    private void complexBusinessLogic() {

        try {
            faultInjectionManager.sleepForAWhile(COMPONENT_NAME);
        } catch (InterruptedException e) {
            LOG.debug("We got a weird exception: {}", e.getMessage());
        }
    }

    private boolean storeRecordsInRemoteStorage() {
        return !faultInjectionManager.maybeFailTheOperation(COMPONENT_NAME);
    }
}
