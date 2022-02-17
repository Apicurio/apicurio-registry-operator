package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.Condition;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;

@Getter @Setter
public class ApicurioRegistryStatus {
    private ArrayList<Condition> conditions;
    private ArrayList<ApicurioRegistryStatusManagedResource> managedResources;
    private ApicurioRegistryStatusInfo info;
}
