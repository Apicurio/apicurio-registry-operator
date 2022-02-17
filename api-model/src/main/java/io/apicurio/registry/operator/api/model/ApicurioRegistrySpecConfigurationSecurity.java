package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationSecurity {
    private ApicurioRegistrySpecConfigurationSecurityKeycloak keycloak;

    public ApicurioRegistrySpecConfigurationSecurityKeycloak getKeycloak() {
        return keycloak;
    }

    public void setKeycloak(ApicurioRegistrySpecConfigurationSecurityKeycloak keycloak) {
        this.keycloak = keycloak;
    }
}
