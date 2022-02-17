package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfiguration {
    private String persistence;
    private ApicurioRegistrySpecConfigurationSql sql;
    private ApicurioRegistrySpecConfigurationKafkasql kafkasql;
    private ApicurioRegistrySpecConfigurationUI ui;
    private String logLevel;
    private ApicurioRegistrySpecConfigurationKafkaSecurity security;

    public String getPersistence() {
        return persistence;
    }

    public void setPersistence(String persistence) {
        this.persistence = persistence;
    }

    public ApicurioRegistrySpecConfigurationSql getSql() {
        return sql;
    }

    public void setSql(ApicurioRegistrySpecConfigurationSql sql) {
        this.sql = sql;
    }

    public ApicurioRegistrySpecConfigurationKafkasql getKafkasql() {
        return kafkasql;
    }

    public void setKafkasql(ApicurioRegistrySpecConfigurationKafkasql kafkasql) {
        this.kafkasql = kafkasql;
    }

    public ApicurioRegistrySpecConfigurationUI getUi() {
        return ui;
    }

    public void setUi(ApicurioRegistrySpecConfigurationUI ui) {
        this.ui = ui;
    }

    public String getLogLevel() {
        return logLevel;
    }

    public void setLogLevel(String logLevel) {
        this.logLevel = logLevel;
    }

    public ApicurioRegistrySpecConfigurationKafkaSecurity getSecurity() {
        return security;
    }

    public void setSecurity(ApicurioRegistrySpecConfigurationKafkaSecurity security) {
        this.security = security;
    }
}
