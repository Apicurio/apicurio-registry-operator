package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfiguration implements KubernetesResource {
    private String persistence;
    private ApicurioRegistrySpecConfigurationSql sql;
    private ApicurioRegistrySpecConfigurationKafkasql kafkasql;
    private ApicurioRegistrySpecConfigurationUI ui;
    private String logLevel;
    private ApicurioRegistrySpecConfigurationKafkaSecurity security;
}
