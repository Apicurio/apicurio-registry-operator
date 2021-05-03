# Prerequisites

1. Install the *Keycloak Operator*
1. Apply the `keycloak.yaml` CR to create a Keycloak instance
1. In development, we can avoid setting up HTTPS:
    - Create an alternate service (see example files)
    - Create a HTTP ingress or route (see example files)
1. Create a realm using the example CR. The realm is configured 
for development environment only.
