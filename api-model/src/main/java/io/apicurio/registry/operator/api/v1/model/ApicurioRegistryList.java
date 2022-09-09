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

import io.fabric8.kubernetes.api.model.ListMeta;
import io.fabric8.kubernetes.client.CustomResourceList;
import io.sundr.builder.annotations.Buildable;
import lombok.ToString;

import java.util.List;
import java.util.Objects;

/**
 * WARNING: The standard equals and hashCode are overridden,
 * so only items are used in the comparison.
 * If full comparison is needed, use the `*Full` methods.
 * Kind and version are not compared.
 */
@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@ToString
// NOTE: We can not use Lombok @Getter and @Setter because it does not work with fabric8 generator.
public class ApicurioRegistryList extends CustomResourceList<ApicurioRegistry> {

    private static final long serialVersionUID = -2979078702023320890L;

    // Dummy field
    @SuppressWarnings("unused")
    private ListMeta metadata;

    // Dummy field
    @SuppressWarnings("unused")
    private List<ApicurioRegistry> items;

    @Override
    public ListMeta getMetadata() {
        return super.getMetadata();
    }

    @Override
    public void setMetadata(ListMeta metadata) {
        super.setMetadata(metadata);
    }

    @Override
    public List<ApicurioRegistry> getItems() {
        return super.getItems();
    }

    @Override
    public void setItems(List<ApicurioRegistry> items) {
        super.setItems(items);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        ApicurioRegistryList other = (ApicurioRegistryList) o;

        return Objects.equals(getItems(), other.getItems());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getItems());
    }

    public boolean equalsFull(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        ApicurioRegistryList other = (ApicurioRegistryList) o;

        return Objects.equals(getMetadata(), other.getMetadata()) &&
                Objects.equals(getItems(), other.getItems());
    }

    public int hashCodeFull() {
        return Objects.hash(getMetadata(),
                getItems());
    }
}
