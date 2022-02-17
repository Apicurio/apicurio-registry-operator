package io.apicurio.registry.operator.api.model;

import io.fabric8.kubernetes.api.model.KubernetesResourceList;
import io.fabric8.kubernetes.api.model.ListMeta;
import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

import java.util.ArrayList;
import java.util.Collection;
import java.util.List;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistryList implements KubernetesResourceList<ApicurioRegistry> {
    private String kind;
    private String apiVersion;

    private ListMeta metadata = new ListMeta();

    private List<ApicurioRegistry> items = new ArrayList<>();

    protected ApicurioRegistryList(final String kind, final String apiVersion) {
        this.kind = kind;
        this.apiVersion = apiVersion;
    }

    public String getKind() {
        return kind;
    }

    public String getApiVersion() {
        return apiVersion;
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

    public void setKind(String kind) {
        this.kind = kind;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public void setItems(List<ApicurioRegistry> items) {
        this.items = items;
    }
}
