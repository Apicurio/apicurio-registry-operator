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

import io.fabric8.kubernetes.api.model.Affinity;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.Toleration;
import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.List;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
@ToString
// NOTE: We can not use Lombok @Getter and @Setter because it does not work with fabric8 generator.
public class ApicurioRegistrySpecDeploymentManagedResources {

    private boolean disableIngress;

    private boolean disableNetworkPolicy;

    private boolean disablePodDisruptionBudget;

    public boolean isDisableIngress() {
        return disableIngress;
    }

    public void setDisableIngress(boolean disableIngress) {
        this.disableIngress = disableIngress;
    }

    public boolean isDisableNetworkPolicy() {
        return disableNetworkPolicy;
    }

    public void setDisableNetworkPolicy(boolean disableNetworkPolicy) {
        this.disableNetworkPolicy = disableNetworkPolicy;
    }

    public boolean isDisablePodDisruptionBudget() {
        return disablePodDisruptionBudget;
    }

    public void setDisablePodDisruptionBudget(boolean disablePodDisruptionBudget) {
        this.disablePodDisruptionBudget = disablePodDisruptionBudget;
    }
}
