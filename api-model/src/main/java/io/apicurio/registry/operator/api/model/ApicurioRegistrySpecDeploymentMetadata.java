package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;

import java.util.HashMap;

public class ApicurioRegistrySpecDeploymentMetadata implements KubernetesResource {
    HashMap<String, String> annotations;
    HashMap<String, String> labels;
}
