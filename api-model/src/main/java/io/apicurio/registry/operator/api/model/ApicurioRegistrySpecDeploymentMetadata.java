package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

import java.util.HashMap;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecDeploymentMetadata {
    HashMap<String, String> annotations;
    HashMap<String, String> labels;

    public HashMap<String, String> getAnnotations() {
        return annotations;
    }

    public void setAnnotations(HashMap<String, String> annotations) {
        this.annotations = annotations;
    }

    public HashMap<String, String> getLabels() {
        return labels;
    }

    public void setLabels(HashMap<String, String> labels) {
        this.labels = labels;
    }
}
