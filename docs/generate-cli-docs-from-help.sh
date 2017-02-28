#!/bin/bash

set -x

declare -a SCS_COMMANDS=("service-registry-info" "service-registry-list" "service-registry-deregister")
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