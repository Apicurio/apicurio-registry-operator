package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Setter @Getter
public class ApicurioRegistrySpecConfigurationSecurity {
    private ApicurioRegistrySpecConfigurationSecurityKeycloak keycloak;
}
