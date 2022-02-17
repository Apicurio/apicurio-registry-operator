package io.apicurio.registry.operator.api.model;


import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;

import java.util.List;
import java.util.Objects;

import static java.util.Collections.singletonList;
import static java.util.Collections.unmodifiableList;

@Version("v1")
@Group("registry.apicur.io")
public class ApicurioRegistry extends CustomResource<ApicurioRegistrySpec, ApicurioRegistryStatus> implements Namespaced {
    public static final String RESOURCE_GROUP = "io.apicurio.registry";
    public static final String RESOURCE_PLURAL = "apicurio-registries";
    public static final String RESOURCE_SINGULAR = "apicurio-registry";
    public static final String CRD_NAME = RESOURCE_PLURAL + "." + RESOURCE_GROUP;
    public static final String SHORT_NAME = "ar";
    public static final List<String> RESOURCE_SHORTNAMES = unmodifiableList(singletonList(SHORT_NAME));

    private String apiVersion;
    private String kind = "ApicurioRegistry";

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

}
