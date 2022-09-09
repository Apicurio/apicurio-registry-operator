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

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;
import lombok.ToString;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
@ToString
// NOTE: We can not use Lombok @Getter and @Setter because it does not work with fabric8 generator.
public class ApicurioRegistrySpecConfigurationKafkaSecurity {

    private ApicurioRegistrySpecConfigurationKafkaSecurityTls tls;

    private ApicurioRegistrySpecConfigurationKafkaSecurityScram scram;

    public ApicurioRegistrySpecConfigurationKafkaSecurityTls getTls() {
        return tls;
    }

    public void setTls(ApicurioRegistrySpecConfigurationKafkaSecurityTls tls) {
        this.tls = tls;
    }

    public ApicurioRegistrySpecConfigurationKafkaSecurityScram getScram() {
        return scram;
    }

    public void setScram(ApicurioRegistrySpecConfigurationKafkaSecurityScram scram) {
        this.scram = scram;
    }
}
