package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.Condition;
import io.fabric8.kubernetes.api.model.KubernetesResource;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;

@Getter @Setter
public class ApicurioRegistryStatus implements KubernetesResource {
    private ArrayList<Condition> conditions;
    private ArrayList<ApicurioRegistryStatusManagedResource> managedResources;
    private ApicurioRegistryStatusInfo info;
}
