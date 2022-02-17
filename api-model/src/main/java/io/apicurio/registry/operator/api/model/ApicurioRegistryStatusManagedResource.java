package io.apicurio.registry.operator.api.model;

import lombok.Getter;
import lombok.Setter;

@Getter @Setter
public class ApicurioRegistryStatusManagedResource {
    private String kind;
    private String name;
    private String namespace;
}
