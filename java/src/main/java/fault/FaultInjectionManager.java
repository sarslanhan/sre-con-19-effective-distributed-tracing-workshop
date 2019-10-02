package fault;

import main.DemoRestTemplateFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import javax.annotation.PostConstruct;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;

@Component
public class FaultInjectionManager {

    private static final Logger LOG = LoggerFactory.getLogger(FaultInjectionManager.class);
    
    private static final long POLLING_INTERVAL = 2 * 1000;
    private static final String[] APP_NAMES = {"super-website", "cart-api"};

    private final Map<String, FaultConfig> configs = new HashMap<>(APP_NAMES.length);

    private final RestTemplate restTemplate = DemoRestTemplateFactory.createDemoRestTemplate();

    @Value("${endpoints.fault}")
    private String getFaultConfigsUrl;
    
    @Value("${ot.az}")
    private int az;

    @PostConstruct
    public void setup() {
        for (String appName : APP_NAMES) {
            FaultConfig config = new FaultConfig();
            config.setAz(this.az);
            config.setErrorRateMax(0);
            config.setLatencyMax(0);
            configs.put(appName, config);
        }
    }
    
    public void sleepForAWhile(final String appName) throws InterruptedException {
        int maxSleepDuration = this.configs.get(appName).getLatencyMax();
        if (maxSleepDuration > 0) {
            int sleepDuration = new Random().nextInt(maxSleepDuration);
            Thread.sleep(sleepDuration);
        }
    }
    
    public boolean maybeFailTheOperation(final String appName) {
        int errorRate = this.configs.get(appName).getErrorRateMax();
        int r = new Random().nextInt(100) - errorRate;
        return r < 0;
    }

    @Scheduled(fixedRate = POLLING_INTERVAL)
    void refreshConfigs() {
        ApplicationFaultConfigTransfer currentConfigs = callFaultInjectionApi();

        for (String appName : APP_NAMES) {
            FaultConfig faultConfig = selectFaultConfig(currentConfigs, appName, this.az);
            if (faultConfig != null) {
                configs.put(appName, faultConfig);
            }
        }
    }

    private FaultConfig selectFaultConfig(final ApplicationFaultConfigTransfer currentConfigs,
                                          final String appName,
                                          final int az) {
        if (!currentConfigs.getConfigs().containsKey(appName)) {
            LOG.debug("Application {} not found", appName);
            return null;
        }

        List<FaultConfig> configs = currentConfigs.getConfigs().get(appName);
        for (FaultConfig config : configs) {
            if (config.getAz() == az) {
                return config;
            }
        }
        LOG.debug("Configuration for app {} and AZ {} not found", appName, az);
        return null;
    }

    private ApplicationFaultConfigTransfer callFaultInjectionApi() {
        ResponseEntity<ApplicationFaultConfigTransfer> currentConfigs = restTemplate.exchange(
                getFaultConfigsUrl,
                HttpMethod.GET,
                null,
                ApplicationFaultConfigTransfer.class);

        if (currentConfigs.getStatusCode().isError()) {
            LOG.warn("Got HTTP response with status code {}", currentConfigs.getStatusCode().value());
            return null;
        }

        if (currentConfigs.getBody().getConfigs().isEmpty()) {
            LOG.warn("HTTP Response from Fault Injection API came back empty");
            return null;
        }

        return currentConfigs.getBody();
    }
}
