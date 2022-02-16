package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationKafkaSecurityTls implements KubernetesResource {
    private String truststoreSecretName;
    private String keystoreSecretName;
}
