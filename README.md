# Spring Cloud Services CLI Plugin

This repository provides a [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) plugin for
Spring Cloud Services.

For information on plugin development, see
[Developing a plugin](https://github.com/cloudfoundry/cli/tree/master/plugin/plugin_examples).

## Go Development

See the [Go Development](docs/go.adoc) guide.
(If you just want to build and install the plugin, simply install go and govendor.)

## Testing

Run the tests as follows:
```bash
$ cd $GOPATH/src/github.com/pivotal-cf/spring-cloud-services-cli-plugin
$ govendor test +local
```

## Building

To build the plugin, install go and govendor (as described above) and issue:
```bash
$ cd $GOPATH/src/github.com/pivotal-cf/spring-cloud-services-cli-plugin
$ govendor install +local
```

## Installing

To install the plugin in the `cf` CLI, first build it and then issue:
```bash
$ cf install-plugin -f $GOPATH/bin/spring-cloud-services-cli-plugin

```

The plugin's commands may then be listed by issuing `cf help`.

## License

The Spring Cloud Services CLI plugin is Open Source software released under the
[Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0.html).

## Contributing

Contributions are welcomed. Please refer to the [Contributor's Guide](CONTRIBUTING.md).
