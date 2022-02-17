package io.apicurio.registry.operator.api.model;


import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpec {
    private ApicurioRegistrySpecConfiguration configuration;
    private ApicurioRegistrySpecDeployment deployment;
}
