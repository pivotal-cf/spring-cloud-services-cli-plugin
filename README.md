# Spring Cloud Services CLI plugin

This repository provides a [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) plugin for
Spring Cloud Services.

For information on plugin development, see
[Developing a plugin](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples).

## Go Development

See the Spring Cloud Services
[Go Development](https://github.com/pivotal-cf/spring-cloud-services-getting-started/blob/master/docs/go.adoc)
guide. (If you just want to build the plugin, you only need to install go and govendor.)

## Building

To build the plugin install go and govendor (as described [here](https://github.com/pivotal-cf/spring-cloud-services-getting-started/blob/master/docs/go.adoc)),
and then issue:
```bash
$ cd $GOPATH/src/github.com/pivotal-cf/spring-cloud-services-cli-plugin
$ govendor install +local
```

## Installing

To install the plugin in the `cf` CLI, first build the plugin and then issue:
```bash
$ cf install-plugin -f $GOPATH/bin/spring-cloud-services-cli-plugin

```

The plugin's commands may then be listed by issuing `cf help`.
