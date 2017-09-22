#!/bin/bash

if [ ! -z ${DEBUG} ]; then
    set -x
fi

declare -a SCS_COMMANDS=("config-server-encrypt-value" "spring-cloud-service-stop" "spring-cloud-service-start" "spring-cloud-service-restart" "service-registry-info" "service-registry-list" "service-registry-enable" "service-registry-deregister" "service-registry-disable")
CMD_DOC_FILENAME=cli.md

echo "# Spring Cloud Services CF CLI Plugin Docs

The following commands can be executed using the Spring Cloud Services [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) Plugin.

# Spring Cloud Services CLI Docs

" > $CMD_DOC_FILENAME

for i in "${SCS_COMMANDS[@]}"
do
    echo "Capturing help documentation from `cf $i` command"
    echo "## \`cf $i\`

\`\`\`" >> $CMD_DOC_FILENAME
    
    cf help $i >> $CMD_DOC_FILENAME
    
    echo "\`\`\`

" >> $CMD_DOC_FILENAME
done

echo "Print contents of $CMD_DOC_FILENAME"
echo "==================================="
cat $CMD_DOC_FILENAME