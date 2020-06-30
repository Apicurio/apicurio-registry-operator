---
layout: default
title: Troubleshooting
nav_order: 5
---

# Troubleshooting
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

TODO

## Managing Environment Variables

UI (OpenShift)
1. Go to the Installed Operators tab, and select the Apicurio Registry operator
2. In the ApicurioRegistry tab, click on the operator CR for your deployment
3. In the main overview page, you'll see the Deployment Name section, which contains the name of DeploymentConfig used for your SR deployment
4. Find that DeploymentConfig by going to the Workloads / Deployment Configs in the left menu
5. Select the DeploymentConfig with the correct name, and switch to the Environment tab
6. You can add your environment variable to the Single values (env) section, and click Save at the bottom

CLI
1. Similarly, select the namespace where Apicurio Registry is installed
2. Use `oc get apicurioregistry` to get the list of ApicurioRegistry CRs and run `oc describe` on the appropriate one
3. You will see Deployment Name in the status section
4. Modity this DeploymentConfig using `oc edit` to add the env. variable

## Liveness and Readiness

Apicurio Registry provides readiness and liveness probes for Kubernetes to ensure application health.
It is set up to use reasonable defaults, but in case users want to adjust them, 
the following is an overview of environment variables that can be used:

- `LIVENESS_ERROR_THRESHOLD` - integer; number of liveness issues/errors that 
  can occur before liveness probe fails
- `LIVENESS_COUNTER_RESET` - seconds; the period in which the *threshold* number of errors 
  must occur, i.e. if this value is 60 and the threshold is 1, the check fails 
  after two errors occur in 1 minute
- `LIVENESS_STATUS_RESET` - seconds; number of seconds that must elapse without any more errors 
  for the liveness probe to reset to *OK* status. 
  *Note: Since Kubernetes restarts the pod that fails the liveness check, 
   this values (unlike readiness) does not actually affect what happens.*
- `LIVENESS_ERRORS_IGNORED` - which is an env. variable containing a comma-separated list 
  of ignored liveness exceptions

- `READINESS_ERROR_THRESHOLD` - integer; number of readiness issues/errors 
  that can occur before readiness probe fails
- `READINESS_COUNTER_RESET` - seconds, same as in the liveness case
- `READINESS_STATUS_RESET` - seconds, same as in the liveness case, 
  in this case it means how long the pod stays not ready, until it returns to normal operation
- `READINESS_TIMEOUT` - seconds, the readiness tracks timeout of two operations - 
  how long it takes for storage request to complete, 
  and how long it takes for HTTP REST API request to return a response. 
  If the operation takes more time than that, it is counted as a readiness issue/error. 
  This value controls those timeouts.

*Definitions:*

 - Liveness - application cannot make progress, Kubernetes will restart the failing pod
 - Readiness - application is not ready (e.g. it's overwhelmed by requests), 
   Kubernetes will stop sending requests for the time the probe fails. 
   If other pods are OK, they will still receive requests.

## Other

TODO
