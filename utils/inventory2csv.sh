#!/bin/bash

jq -r '
  ["name", "artifactId", "url", "licenses_name"],
  (.libraries[] |
    [
      .name,
      .artifactId,
      .references.url,
      ([.licenses[].name] | join(","))
    ]
  )
| @csv
' $1 > $2
