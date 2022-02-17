package io.apicurio.registry.operator.api.model;


import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpec {
    private ApicurioRegistrySpecConfiguration configuration;
    private ApicurioRegistrySpecDeployment deployment;

    public ApicurioRegistrySpecConfiguration getConfiguration() {
        return configuration;
    }

    public void setConfiguration(ApicurioRegistrySpecConfiguration configuration) {
        this.configuration = configuration;
    }

    public ApicurioRegistrySpecDeployment getDeployment() {
        return deployment;
    }

    public void setDeployment(ApicurioRegistrySpecDeployment deployment) {
        this.deployment = deployment;
    }
}
