package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Setter @Getter
public class ApicurioRegistrySpecConfigurationKafkaSecurity {
    private ApicurioRegistrySpecConfigurationKafkaSecurityTls tls;
    private ApicurioRegistrySpecConfigurationKafkaSecurityScram scram;
}
