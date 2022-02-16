package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.Affinity;
import io.fabric8.kubernetes.api.model.KubernetesResource;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.Toleration;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;

@Setter @Getter
public class ApicurioRegistrySpecDeployment implements KubernetesResource {
    private Integer replicas;
    private String host;
    private Affinity affinity;
    private ArrayList<Toleration> tolerations;
    private ApicurioRegistrySpecDeploymentMetadata metadata;
    private String image;
    private ArrayList<LocalObjectReference> imagePullSecrets;
}
