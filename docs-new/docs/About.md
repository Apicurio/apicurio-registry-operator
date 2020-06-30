---
layout: default
title: About
nav_order: 6
---

# About
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Operator-Operand Version Table

When user runs an operator, it has to know which version of the operand to deploy 
and how to manage it. In Apicurio Registry Operator, this information is provided 
as a set of environment variables in the operator's `Deployment` resource.
 
Operator and operand are usually projects that evolve together, but their code is kept separate. 
If a bug is discovered in an operand, a new version will be released to fix the issue.
However, the operator code does not necessarily change, it only has to know that a newer version 
of the operand should be deployed. This situation also applies in reverse.    
 
Some operators support a fixed operand versioning, 
so a new operand release requires a release of a new operator version. 
The approach we use for the Apicurio Registry Operator is a bit different.
Operator versions are loosely decoupled from the operand versions, 
so given version of the operator may support several different operand versions, and vice versa.
(This is usually the case for micro versions.)

To keep track of this relationship, the following table may be useful:

|Operator|Operand (`alpha` channel)|
|---|---|
|0.0.3|1.2.3.Final|
|0.0.2 and previous|-|

One of the easiest way to deploy and manage Operators for users is to use the [Operator Lifecycle Manager (OLM)](https://docs.openshift.com/container-platform/latest/operators/understanding_olm/olm-understanding-olm.html).
Apicurio Registry Operator is released in the Operator Hub, which takes care of providing the correct information to deploy it. 
This information is packaged in a set of *Cluster Service Version (CVS)* artifacts, which keep track of the correct
Operator-Operand versions to use.

In Apicurio Registry Operator, we started naming these to reflect both versions. 
For example, a `0.0.3-v1.2.3.final` CSV will install a `0.0.3` version of the operator,
which deploys Apicurio Registry `1.2.3.Final`.  

To better organize the CSVs, they are grouped together in a *channel*. 
A channel is an ordered list of CSVs, that enable the OLM to determine how to upgrade the operator (and operands).

Currently, we publish in the `alpha` channel, but in the future we will use a channel 
for each major (or incompatible) version of the Registry.

**Note**: The Operator is still in the alpha version stage. If you subscribe to the`alpha` channel, it is not recommended to use automatic updates, 
so you can test if it's applied correctly.

## Release Notes

### 0.0.4 (dev)

### 0.0.3 (latest)
 
 - TODO
  
### 0.0.2 and previous

- TODO
