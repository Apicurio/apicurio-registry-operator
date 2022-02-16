package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpecConfigurationSecurityKeycloak implements KubernetesResource {
    private String url;
    private String realm;
    private String apiClientId;
    private String uiClientId;
}
