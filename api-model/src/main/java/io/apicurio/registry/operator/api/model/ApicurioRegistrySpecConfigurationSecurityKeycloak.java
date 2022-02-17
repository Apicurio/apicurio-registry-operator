package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationSecurityKeycloak {
    private String url;
    private String realm;
    private String apiClientId;
    private String uiClientId;
}
