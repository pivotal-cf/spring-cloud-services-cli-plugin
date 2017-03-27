# Spring Cloud Services CF CLI Plugin Docs

The following commands can be executed using the Spring Cloud Services [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) Plugin.

# Spring Cloud Services CLI Docs


## `cf config-server-encrypt-value`

```
NAME:
   config-server-encrypt-value - Encrypt a string using a Spring Cloud Services configuration server

USAGE:
      cf config-server-encrypt-value CONFIG_SERVER_INSTANCE_NAME VALUE_TO_ENCRYPT

ALIAS:
   csev

OPTIONS:
   --skip-ssl-validation      Skip verification of the service endpoint. Not recommended!
```


## `cf service-registry-info`

```
NAME:
   service-registry-info - Display Spring Cloud Services service registry instance information

USAGE:
      cf service-registry-info SERVICE_REGISTRY_INSTANCE_NAME

ALIAS:
   sri

OPTIONS:
   --skip-ssl-validation      Skip verification of the service endpoint. Not recommended!
```


## `cf service-registry-list`

```
NAME:
   service-registry-list - Display all applications registered with a Spring Cloud Services service registry

USAGE:
      cf service-registry-list SERVICE_REGISTRY_INSTANCE_NAME

ALIAS:
   srl

OPTIONS:
   --skip-ssl-validation      Skip verification of the service endpoint. Not recommended!
```


## `cf service-registry-deregister`

```
NAME:
   service-registry-deregister - Deregister an application registered with a Spring Cloud Services service registry

USAGE:
      cf service-registry-deregister SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME

ALIAS:
   srd

OPTIONS:
   --skip-ssl-validation        Skip verification of the service endpoint. Not recommended!
   --i/--cf-instance-index      Deregister a specific instance in the Eureka registry. The instance index number can be found by using the the service-registry-list command.
```


