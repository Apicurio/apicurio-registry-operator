package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResource;
import io.fabric8.kubernetes.api.model.KubernetesResourceList;
import io.fabric8.kubernetes.api.model.ListMeta;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;

@Setter @Getter
public class ApicurioRegistryList implements KubernetesResource, KubernetesResourceList<ApicurioRegistry> {
    private String kind;
    private String apiVersion;

    private ListMeta metadata = new ListMeta();

    private List<ApicurioRegistry> items = new ArrayList<>();

    protected ApicurioRegistryList(final String kind, final String apiVersion) {
        this.kind = kind;
        this.apiVersion = apiVersion;
    }

    public void setItems(final Collection<ApicurioRegistry> items) {
        this.items = new ArrayList<>(items);
    }

    public List<ApicurioRegistry> getItems() {
        return this.items;
    }

    public void setMetadata(final ListMeta metadata) {
        this.metadata = metadata;
    }

    @Override
    public ListMeta getMetadata() {
        return this.metadata;
    }

    @Override
    public String toString() {
        return "{metadata=" + this.metadata + "," +
                "items=" + this.items + "}";
    }
}
