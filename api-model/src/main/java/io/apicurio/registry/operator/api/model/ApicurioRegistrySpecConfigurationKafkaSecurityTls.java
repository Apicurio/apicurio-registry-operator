package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationKafkaSecurityTls {
    private String truststoreSecretName;
    private String keystoreSecretName;
}
