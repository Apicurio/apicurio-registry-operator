package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.Affinity;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.Toleration;
import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

import java.util.ArrayList;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecDeployment {
    private Integer replicas;
    private String host;
    private Affinity affinity;
    private ArrayList<Toleration> tolerations;
    private ApicurioRegistrySpecDeploymentMetadata metadata;
    private String image;
    private ArrayList<LocalObjectReference> imagePullSecrets;

    public Integer getReplicas() {
        return replicas;
    }

    public void setReplicas(Integer replicas) {
        this.replicas = replicas;
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public Affinity getAffinity() {
        return affinity;
    }

    public void setAffinity(Affinity affinity) {
        this.affinity = affinity;
    }

    public ArrayList<Toleration> getTolerations() {
        return tolerations;
    }

    public void setTolerations(ArrayList<Toleration> tolerations) {
        this.tolerations = tolerations;
    }

    public ApicurioRegistrySpecDeploymentMetadata getMetadata() {
        return metadata;
    }

    public void setMetadata(ApicurioRegistrySpecDeploymentMetadata metadata) {
        this.metadata = metadata;
    }

    public String getImage() {
        return image;
    }

    public void setImage(String image) {
        this.image = image;
    }

    public ArrayList<LocalObjectReference> getImagePullSecrets() {
        return imagePullSecrets;
    }

    public void setImagePullSecrets(ArrayList<LocalObjectReference> imagePullSecrets) {
        this.imagePullSecrets = imagePullSecrets;
    }
}
