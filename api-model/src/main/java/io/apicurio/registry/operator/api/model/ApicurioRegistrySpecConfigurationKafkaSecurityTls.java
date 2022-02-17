package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationKafkaSecurityTls {
    private String truststoreSecretName;
    private String keystoreSecretName;

    public String getTruststoreSecretName() {
        return truststoreSecretName;
    }

    public void setTruststoreSecretName(String truststoreSecretName) {
        this.truststoreSecretName = truststoreSecretName;
    }

    public String getKeystoreSecretName() {
        return keystoreSecretName;
    }

    public void setKeystoreSecretName(String keystoreSecretName) {
        this.keystoreSecretName = keystoreSecretName;
    }
}
