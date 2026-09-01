module wss

go 1.24.0

require (
	github.com/joho/godotenv v1.5.1
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/text v0.28.0

replace (
	github.com/joho/godotenv v1.5.1 => ./projectPackages/github.com/joho/godotenv
	golang.org/x/mod v0.26.0 => ./projectPackages/golang.org/x/mod
	golang.org/x/sync v0.16.0 => ./projectPackages/golang.org/x/sync
	golang.org/x/text v0.28.0 => ./projectPackages/golang.org/x/text
	golang.org/x/tools v0.35.0 => ./projectPackages/golang.org/x/tools
	gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405 => ./projectPackages/gopkg.in/check.v1
	gopkg.in/yaml.v3 v3.0.1 => ./projectPackages/gopkg.in/yaml.v3
)
