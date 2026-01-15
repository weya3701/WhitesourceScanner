#!/bin/bash

sed -i "" 's/\.pom/\.jar/g' $1

jq -r '
  ["name", "artifactId", "url", "licenses_name"],
  (.libraries[] |
    [
      .name,
      .artifactId,
      .references.pomUrl,
      ([.licenses[].name] | join(","))
    ]
  )
| @csv
' $1 > $2
