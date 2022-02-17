package io.apicurio.registry.operator.api.model;

import io.sundr.builder.annotations.Buildable;
import lombok.EqualsAndHashCode;

@Buildable(
        editableEnabled = false,
        builderPackage = Constants.FABRIC8_KUBERNETES_API
)
@EqualsAndHashCode
public class ApicurioRegistrySpecConfigurationKafkaSecurityScram {
    String trustedSecretName;
    String user;
    String passwordSecretName;
    String mechanism;

    public String getTrustedSecretName() {
        return trustedSecretName;
    }

    public void setTrustedSecretName(String trustedSecretName) {
        this.trustedSecretName = trustedSecretName;
    }

    public String getUser() {
        return user;
    }

    public void setUser(String user) {
        this.user = user;
    }

    public String getPasswordSecretName() {
        return passwordSecretName;
    }

    public void setPasswordSecretName(String passwordSecretName) {
        this.passwordSecretName = passwordSecretName;
    }

    public String getMechanism() {
        return mechanism;
    }

    public void setMechanism(String mechanism) {
        this.mechanism = mechanism;
    }
}
