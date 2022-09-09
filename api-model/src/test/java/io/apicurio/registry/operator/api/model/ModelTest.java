package io.apicurio.registry.operator.api.model;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import io.apicurio.registry.operator.api.v1.model.ApicurioRegistry;
import io.apicurio.registry.operator.api.v1.model.ApicurioRegistryBuilder;
import io.apicurio.registry.operator.api.v1.model.ApicurioRegistryList;
import io.apicurio.registry.operator.api.v1.model.ApicurioRegistryListBuilder;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

class ModelTest {

    public static final ObjectMapper MAPPER;

    static {
        MAPPER = new ObjectMapper(new YAMLFactory());
    }

    private final Logger log = LoggerFactory.getLogger(getClass());

    @Test
    void basicSerDesTest() throws IOException {

        var ar1 = new ApicurioRegistryBuilder()
                .withNewMetadata()
                    .withName("test")
                    .withNamespace("test-namespace")
                .endMetadata()
                .withNewSpec()
                    .withNewConfiguration()
                        .withPersistence("mem")
                        .withEnv(new EnvVarBuilder()
                                .withName("VAR_1_NAME")
                                .withValue("VAR_1_VALUE")
                                .build()
                        )
                    .endConfiguration()
                .endSpec()
                .build();

        var ar2 = MAPPER.readValue(
                getClass().getResourceAsStream("/apicurio-registry-cr.yaml"),
                ApicurioRegistry.class);

        Assertions.assertEquals(ar1, ar2);
        ar1.getSpec().getConfiguration().getEnv().get(0).setName("VAR_2_NAME");
        Assertions.assertNotEquals(ar1, ar2);
        ar1.getSpec().getConfiguration().getEnv().get(0).setName("VAR_1_NAME");

        // LIST

        var arl1 = new ApicurioRegistryListBuilder()
                .withItems(ar1)
                .build();

        var arl2 = MAPPER.readValue(
                getClass().getResourceAsStream("/apicurio-registry-cr-list.yaml"),
                ApicurioRegistryList.class);

        Assertions.assertEquals(arl1, arl2);
    }
}
