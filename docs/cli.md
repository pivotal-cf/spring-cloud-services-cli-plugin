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
```


## `cf scs-stop`

```
NAME:
   spring-cloud-service-stop - Stop a Spring Cloud Services service instance

USAGE:
      cf scs-stop SERVICE_INSTANCE_NAME

ALIAS:
   scs-stop
```

## `cf scs-start`

```
NAME:
   spring-cloud-service-start - Start a Spring Cloud Services service instance

USAGE:
      cf scs-start SERVICE_INSTANCE_NAME

ALIAS:
   scs-start
```

## `cf scs-restart`

```
NAME:
   spring-cloud-service-restart - Restart a Spring Cloud Services service instance

USAGE:
      cf scs-restart SERVICE_INSTANCE_NAME

ALIAS:
   scs-restart
```

## `cf scs-restage`

```
NAME:
   spring-cloud-service-restage - Restage a Spring Cloud Services service instance

USAGE:
      cf scs-restage SERVICE_INSTANCE_NAME

ALIAS:
   scs-restage
```


## `cf service-registry-info`

```
NAME:
   service-registry-info - Display Spring Cloud Services service registry instance information

USAGE:
      cf service-registry-info SERVICE_REGISTRY_INSTANCE_NAME

ALIAS:
   sri
```


## `cf service-registry-list`

```
NAME:
   service-registry-list - Display all applications registered with a Spring Cloud Services service registry

USAGE:
      cf service-registry-list SERVICE_REGISTRY_INSTANCE_NAME

ALIAS:
   srl
```


## `cf service-registry-enable`

```
NAME:
   service-registry-enable - Enable an application registered with a Spring Cloud Services service registry so that it is available for traffic

USAGE:
      cf service-registry-enable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME

ALIAS:
   sren

OPTIONS:
   --i/--cf-instance-index      Operate on a specific instance in the Eureka registry. The instance index number can be found by using the service-registry-list command.
```


## `cf service-registry-deregister`

```
NAME:
   service-registry-deregister - Deregister an application registered with a Spring Cloud Services service registry

USAGE:
      cf service-registry-deregister SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME

ALIAS:
   srdr

OPTIONS:
   --i/--cf-instance-index      Operate on a specific instance in the Eureka registry. The instance index number can be found by using the service-registry-list command.
```


## `cf service-registry-disable`

```
NAME:
   service-registry-disable - Disable an application registered with a Spring Cloud Services service registry so that it is unavailable for traffic

USAGE:
      cf service-registry-disable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME

ALIAS:
   srda

OPTIONS:
   --i/--cf-instance-index      Operate on a specific instance in the Eureka registry. The instance index number can be found by using the service-registry-list command.
```


