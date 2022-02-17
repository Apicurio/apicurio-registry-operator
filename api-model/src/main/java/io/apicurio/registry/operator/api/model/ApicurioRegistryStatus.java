package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.Condition;
import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

import java.util.ArrayList;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistryStatus {
    private ArrayList<Condition> conditions;
    private ArrayList<ApicurioRegistryStatusManagedResource> managedResources;
    private ApicurioRegistryStatusInfo info;

    public ArrayList<Condition> getConditions() {
        return conditions;
    }

    public void setConditions(ArrayList<Condition> conditions) {
        this.conditions = conditions;
    }

    public ArrayList<ApicurioRegistryStatusManagedResource> getManagedResources() {
        return managedResources;
    }

    public void setManagedResources(ArrayList<ApicurioRegistryStatusManagedResource> managedResources) {
        this.managedResources = managedResources;
    }

    public ApicurioRegistryStatusInfo getInfo() {
        return info;
    }

    public void setInfo(ApicurioRegistryStatusInfo info) {
        this.info = info;
    }
}
