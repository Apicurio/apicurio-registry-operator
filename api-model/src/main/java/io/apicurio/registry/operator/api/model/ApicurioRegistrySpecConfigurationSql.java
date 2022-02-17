package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.Setter;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationSql {
    ApicurioRegistrySpecConfigurationDataSource dataSource;

    public ApicurioRegistrySpecConfigurationDataSource getDataSource() {
        return dataSource;
    }

    public void setDataSource(ApicurioRegistrySpecConfigurationDataSource dataSource) {
        this.dataSource = dataSource;
    }
}
