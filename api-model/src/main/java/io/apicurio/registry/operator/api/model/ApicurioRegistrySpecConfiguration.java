package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfiguration {
    private String persistence;
    private ApicurioRegistrySpecConfigurationSql sql;
    private ApicurioRegistrySpecConfigurationKafkasql kafkasql;
    private ApicurioRegistrySpecConfigurationUI ui;
    private String logLevel;
    private ApicurioRegistrySpecConfigurationKafkaSecurity security;
}
