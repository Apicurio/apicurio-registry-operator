package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistryStatusManagedResource implements KubernetesResource {
    private String kind;
    private String name;
    private String namespace;
}
