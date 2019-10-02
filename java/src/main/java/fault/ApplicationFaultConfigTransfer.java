package fault;

import java.io.Serializable;
import java.util.List;
import java.util.Map;

public class ApplicationFaultConfigTransfer implements Serializable {
    private Map<String, List<FaultConfig>> configs;

    public Map<String, List<FaultConfig>> getConfigs() {
        return configs;
    }

    public void setConfigs(final Map<String, List<FaultConfig>> configs) {
        this.configs = configs;
    }
}
