package io.apicurio.registry.operator.api.model;


import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;
import io.sundr.builder.annotations.Buildable;
import io.sundr.builder.annotations.BuildableReference;

import java.util.Objects;


@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API,
        refs = {@BuildableReference(ObjectMeta.class)}
)
@Version(Constants.API_VERSION)
@Group(Constants.RESOURCE_GROUP)
public class ApicurioRegistry extends CustomResource<ApicurioRegistrySpec, ApicurioRegistryStatus> implements Namespaced {
    private ObjectMeta metadata; // Leave these attributes for generator
    private ApicurioRegistrySpec spec;
    private ApicurioRegistryStatus status;

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder("{");
        sb.append("metadata=").append(getMetadata()).append(",");
        sb.append("spec=").append(spec).append(",");
        sb.append("status=").append(status);
        sb.append("}");
        return sb.toString();
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        ApicurioRegistry other = (ApicurioRegistry) o;

        return Objects.equals(getMetadata().getNamespace(), other.getMetadata().getNamespace()) &&
                Objects.equals(spec, other.spec);
    }

    @Override
    public int hashCode() {
        return Objects.hash(getMetadata().getName(), spec );
    }

    @Override
    public ApicurioRegistrySpec getSpec() {
        return this.spec;
    }

    @Override
    public void setSpec(ApicurioRegistrySpec spec) {
        this.spec = spec;
    }

    @Override
    public ApicurioRegistryStatus getStatus() {
        return this.status;
    }

    @Override
    public void setStatus(ApicurioRegistryStatus status) {
        this.status = status;
    }
}
