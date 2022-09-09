/*
 * Copyright 2022 Red Hat
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.apicurio.registry.operator.api.v1.model;

import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;
import io.sundr.builder.annotations.Buildable;
import io.sundr.builder.annotations.BuildableReference;
import lombok.ToString;

import java.util.Objects;

/**
 * WARNING: The standard equals and hashCode are overridden,
 * so only namespace, name and spec are used in the comparison.
 * If full comparison is needed, use the `*Full` methods.
 * Kind and version are not compared.
 */
@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API,
        refs = {@BuildableReference(ObjectMeta.class)}
)
@Version(Constants.API_VERSION)
@Group(Constants.RESOURCE_GROUP)
@ToString
// NOTE: We can not use Lombok @Getter and @Setter because it does not work with fabric8 generator.
public class ApicurioRegistry extends CustomResource<ApicurioRegistrySpec, ApicurioRegistryStatus> implements Namespaced {

    private static final long serialVersionUID = 2047271075581318563L;

    // Dummy field
    @SuppressWarnings("unused")
    private ObjectMeta metadata;

    // Dummy field
    @SuppressWarnings("unused")
    private ApicurioRegistrySpec spec;

    // Dummy field
    @SuppressWarnings("unused")
    private ApicurioRegistrySpec status;

    @Override
    public ObjectMeta getMetadata() {
        return super.getMetadata();
    }

    @Override
    public void setMetadata(ObjectMeta metadata) {
        super.setMetadata(metadata);
    }

    @Override
    public ApicurioRegistrySpec getSpec() {
        return super.getSpec();
    }

    @Override
    public void setSpec(ApicurioRegistrySpec spec) {
        super.setSpec(spec);
    }

    @Override
    public ApicurioRegistryStatus getStatus() {
        return super.getStatus();
    }

    @Override
    public void setStatus(ApicurioRegistryStatus status) {
        super.setStatus(status);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        ApicurioRegistry other = (ApicurioRegistry) o;

        return Objects.equals(getMetadata().getNamespace(), other.getMetadata().getNamespace()) &&
                Objects.equals(getMetadata().getName(), other.getMetadata().getName()) &&
                Objects.equals(getSpec(), other.getSpec());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getMetadata().getNamespace(),
                getMetadata().getName(),
                getSpec());
    }

    public boolean equalsFull(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        ApicurioRegistry other = (ApicurioRegistry) o;

        return Objects.equals(getMetadata(), other.getMetadata()) &&
                Objects.equals(getSpec(), other.getSpec()) &&
                Objects.equals(getStatus(), other.getStatus());
    }

    public int hashCodeFull() {
        return Objects.hash(getMetadata(),
                getSpec(),
                getStatus());
    }
}
