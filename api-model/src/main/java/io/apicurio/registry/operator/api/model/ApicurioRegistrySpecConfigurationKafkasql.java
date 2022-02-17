package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationKafkasql {
    private String bootstrapServers;
    ApicurioRegistrySpecConfigurationKafkaSecurity security;
}
