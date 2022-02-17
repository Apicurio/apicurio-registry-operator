package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationKafkaSecurity {
    private ApicurioRegistrySpecConfigurationKafkaSecurityTls tls;
    private ApicurioRegistrySpecConfigurationKafkaSecurityScram scram;

    public ApicurioRegistrySpecConfigurationKafkaSecurityTls getTls() {
        return tls;
    }

    public void setTls(ApicurioRegistrySpecConfigurationKafkaSecurityTls tls) {
        this.tls = tls;
    }

    public ApicurioRegistrySpecConfigurationKafkaSecurityScram getScram() {
        return scram;
    }

    public void setScram(ApicurioRegistrySpecConfigurationKafkaSecurityScram scram) {
        this.scram = scram;
    }
}
