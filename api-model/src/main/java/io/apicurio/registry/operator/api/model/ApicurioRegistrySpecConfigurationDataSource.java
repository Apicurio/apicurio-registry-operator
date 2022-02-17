package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationDataSource {
    private String url;
    private String username;
    private String password;
}
