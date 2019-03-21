# Spring Cloud Services CLI Plugin

This repository provides a [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) plugin for
Spring Cloud Services.

For information on plugin development, see
[Developing a plugin](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples).

## Building

To build the plugin, install go and govendor (see the [Go Development](docs/go.adoc) guide for instructions) and issue:
```bash
$ rm $GOPATH/bin/spring-cloud-services-cli-plugin
$ cd $GOPATH/src/github.com/pivotal-cf/spring-cloud-services-cli-plugin
$ govendor install -ldflags="-X main.pluginVersion=$(cat version)" +local
```
This builds the plugin with the current version number in the [version file](version).

Note: if an invalid version number is provided, the build will succeed, but the plugin will fail to install (with exit status code 64).

To print the version number of the built plugin, run it as a stand-alone executable, for example:
```bash
$ $GOPATH/bin/spring-cloud-services-cli-plugin
This program is a plugin which expects to be installed into the cf CLI. It is not intended to be run stand-alone.
Plugin version: 0.0.8
```

## Installing

To install the plugin in the `cf` CLI, first build it and then issue:
```bash
$ cf install-plugin -f $GOPATH/bin/spring-cloud-services-cli-plugin

```

You can also install the plugin from the [Cloud Foundry cf-cli plugins repository](https://plugins.cloudfoundry.org):
```bash
$ cf install-plugin -r CF-Community "spring-cloud-services"
```

The plugin's commands may then be listed by issuing `cf help`.

To update the plugin, uninstall it as follows and then re-install the plugin as above:
```bash
$ cf uninstall-plugin spring-cloud-services
```

## Command docs

The Spring Cloud Services CLI plugin command docs can be generated by running the following commands:

```
$ cd docs
$ ./generate-cli-docs-from-help.bash
```

This needs to be done whenever commands are added, modified, or deleted. Note that the script contains a list of commands which needs to be kept in step with the available commands.

The generated docs may be viewed [here](docs/cli.md).

## Go Development

See the [Go Development](docs/go.adoc) guide.
(If you just want to build and install the plugin, simply install go and govendor.)

## Testing

Run the tests as follows:
```bash
$ cd $GOPATH/src/github.com/pivotal-cf/spring-cloud-services-cli-plugin
$ govendor test +local
```

## License

The Spring Cloud Services CLI plugin is Open Source software released under the
[Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0.html).

## Contributing

Contributions are welcomed. Please refer to the [Contributor's Guide](CONTRIBUTING.md).
