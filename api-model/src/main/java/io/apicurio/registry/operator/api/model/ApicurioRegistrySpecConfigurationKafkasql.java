package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationKafkasql {
    private String bootstrapServers;
    ApicurioRegistrySpecConfigurationKafkaSecurity security;

    public String getBootstrapServers() {
        return bootstrapServers;
    }

    public void setBootstrapServers(String bootstrapServers) {
        this.bootstrapServers = bootstrapServers;
    }

    public ApicurioRegistrySpecConfigurationKafkaSecurity getSecurity() {
        return security;
    }

    public void setSecurity(ApicurioRegistrySpecConfigurationKafkaSecurity security) {
        this.security = security;
    }
}
