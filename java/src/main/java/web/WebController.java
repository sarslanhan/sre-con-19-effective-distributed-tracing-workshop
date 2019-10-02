package web;

import fault.FaultInjectionManager;
import io.opentracing.Span;
import io.opentracing.SpanContext;
import io.opentracing.Tracer;
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
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Collectors;

import static io.opentracing.propagation.Format.Builtin.HTTP_HEADERS;

@Controller
public class WebController {
    private static final Logger LOG = LoggerFactory.getLogger(WebController.class);
    
    private static final String COMPONENT_NAME = "super-website";
    private static final String OPERATION_NAME = "buy_stuff";
    private static final String BUSINESS_LOGIC_OPERATION_NAME = "do_complex_business_logic";
    private static final String RECORD_STORAGE_OPERATION_NAME = "store_records";
    private static final String ADD_TO_CART_OPERATION_NAME = "add_to_cart";
    private static final String MISSING_SKU = "missing sku";
    private static final String ITEM_NOT_ADDED = "sku could not be added to cart";
    private static final String ITEM_ADDED = "item added to cart";
    private static final String CART_API_ERROR = "error calling Cart API";
    private static final String REMOTE_STORAGE_FAILURE = "failed to add record to remote storage";
    private static final String MODEL_ATTRIBUTE_MESSAGE = "message";
    private static final String MODEL_ATTRIBUTE_SKU = "sku";

    private final Tracer tracer;

    @Value("${endpoints.cart}")
    private String cartEndpoint;

    @Value("${ot.az}")
    private int az;

    @Value("${ot.instance_id}")
    private String instanceId;
    
    private final FaultInjectionManager faultInjectionManager;

    private final RestTemplate restTemplate = DemoRestTemplateFactory.createDemoRestTemplate();

    @Autowired
    public WebController(final FaultInjectionManager faultInjectionManager, final TracingFactory tracingFactory){
        this.faultInjectionManager = faultInjectionManager;
        this.tracer = tracingFactory.createTracer(COMPONENT_NAME);
    }

    @RequestMapping("/")
    public String cart() {
        return "cart";
    }

    @RequestMapping(value = "/buy", method = RequestMethod.POST)
    public String buyItems(@RequestParam(value = "sku") String sku, ModelMap model, HttpServletRequest request) {

        if (sku.isEmpty()) {
            model.addAttribute(MODEL_ATTRIBUTE_MESSAGE, MISSING_SKU);
            return "buy";
        }

        complexBusinessLogic();

        boolean success = storeRecordsInRemoteStorage();
        if (!success) {
            model.addAttribute(MODEL_ATTRIBUTE_MESSAGE, REMOTE_STORAGE_FAILURE);
            return "buy";
        }

        model.addAttribute(MODEL_ATTRIBUTE_SKU, sku);

        String url = cartEndpoint + sku;
        ResponseEntity<Boolean> cartResponse = restTemplate.exchange(url,
                HttpMethod.POST,
                null,
                Boolean.class);

        if (cartResponse.getStatusCode().isError()) {
            model.addAttribute(MODEL_ATTRIBUTE_MESSAGE, ITEM_NOT_ADDED);
            return "buy";
        }

        if (cartResponse.getBody()) {
            model.addAttribute(MODEL_ATTRIBUTE_MESSAGE, ITEM_ADDED);
        } else {
            model.addAttribute(MODEL_ATTRIBUTE_MESSAGE, ITEM_NOT_ADDED);
        }
        return "buy";
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
