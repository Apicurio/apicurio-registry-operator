package io.apicurio.registry.operator.api.model;


import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistrySpec implements KubernetesResource {
    private ApicurioRegistrySpecConfiguration configuration;
    private ApicurioRegistrySpecDeployment deployment;
}
