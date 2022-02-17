package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Setter @Getter
public class ApicurioRegistrySpecConfigurationKafkaSecurityScram {
    String trustedSecretName;
    String user;
    String passwordSecretName;
    String mechanism;
}
