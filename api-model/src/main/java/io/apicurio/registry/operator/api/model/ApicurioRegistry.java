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

package io.apicurio.registry.operator.api.model;


import java.util.Objects;

import io.fabric8.kubernetes.api.model.Namespaced;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.client.CustomResource;
import io.fabric8.kubernetes.model.annotation.Group;
import io.fabric8.kubernetes.model.annotation.Version;
import io.sundr.builder.annotations.Buildable;
import io.sundr.builder.annotations.BuildableReference;


@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API,
        refs = {@BuildableReference(ObjectMeta.class)}
)
@Version(Constants.API_VERSION)
@Group(Constants.RESOURCE_GROUP)
public class ApicurioRegistry extends CustomResource<ApicurioRegistrySpec, ApicurioRegistryStatus> implements Namespaced {

    private static final long serialVersionUID = 2047271075581318563L;

    @SuppressWarnings("unused")
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
