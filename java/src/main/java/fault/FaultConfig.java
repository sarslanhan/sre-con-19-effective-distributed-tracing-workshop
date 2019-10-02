package fault;

import java.io.Serializable;

public class FaultConfig implements Serializable {

    private int latencyMax;
    private int errorRateMax;
    private int az;

    public int getLatencyMax() {
        return latencyMax;
    }

    public void setLatencyMax(final int latencyMax) {
        this.latencyMax = latencyMax;
    }

    public int getErrorRateMax() {
        return errorRateMax;
    }

    public void setErrorRateMax(final int errorRateMax) {
        this.errorRateMax = errorRateMax;
    }

    public int getAz() {
        return az;
    }

    public void setAz(final int az) {
        this.az = az;
    }
}
