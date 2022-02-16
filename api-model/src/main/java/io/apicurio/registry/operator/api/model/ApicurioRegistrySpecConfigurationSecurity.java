package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Setter @Getter
public class ApicurioRegistrySpecConfigurationSecurity implements KubernetesResource {
    private ApicurioRegistrySpecConfigurationSecurityKeycloak keycloak;
}
