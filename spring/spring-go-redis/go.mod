module github.com/go-spring/spring-go-redis

go 1.14

require (
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-spring/spring-core v1.1.0-rc1
)

//replace github.com/go-spring/spring-core => ../spring-core
//replace github.com/go-spring/spring-base => ../spring-base