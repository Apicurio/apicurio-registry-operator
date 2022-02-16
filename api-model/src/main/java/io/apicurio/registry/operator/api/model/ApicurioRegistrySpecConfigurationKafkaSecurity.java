package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Setter @Getter
public class ApicurioRegistrySpecConfigurationKafkaSecurity implements KubernetesResource {
    private ApicurioRegistrySpecConfigurationKafkaSecurityTls tls;
    private ApicurioRegistrySpecConfigurationKafkaSecurityScram scram;
}
