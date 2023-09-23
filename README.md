# Go-Spring

[![GoDoc][1]][2] [![license-Apache 2][3]][4]

<!--[![Downloads][7]][8]-->

[1]: https://godoc.org/github.com/go-spring-projects/go-spring?status.svg
[2]: https://godoc.org/github.com/go-spring-projects/go-spring
[3]: https://img.shields.io/badge/license-Apache%202-blue.svg
[4]: LICENSE

`Go-Spring` vision is to empower Go programmers with a powerful programming framework similar to Java `Spring`. It is dedicated to providing users with a simple, secure, and reliable programming experience.

This project based from [go-spring/go-spring](https://github.com/go-spring/go-spring) created by [lvan100](https://github.com/lvan100)
* Switch to monolithic repository.
* Remove third-party modules and retain only the core dependency injection functionality.
* Invoke `AppRunner` and `AppEvent` in the order of their dependencies.
* Add the dynamic property types like `Array`/`Map`/`Value`.
* Add named logger.

### Install
`go get github.com/go-spring-projects/go-spring@latest`

### IoC container

In addition to implementing a powerful IoC container similar to Java Spring, Go-Spring also extends the concept of beans. In Go, objects (pointers), arrays, maps, and function pointers can all be considered beans and can be placed in the IoC container.

| Java Spring 				                      | Go-Spring			                   |
|:--------------------------------------|:-------------------------------|
| `@Value` 								                     | `value:"${}"` 				             |
| `@Autowired` `@Qualifier` `@Required` | `autowire:"?"` 				            |
| `@Configurable` 						                | `WireBean()` 					             |
| `@Profile` 							                    | `ConditionOnProfile()` 		      |
| `@Primary` 							                    | `Primary()` 					              |
| `@DependsOn` 							                  | `DependsOn()` 				             |
| `@ConstructorBinding` 				            | `RegisterBeanFn()` 			         |
| `@ComponentScan` `@Indexed` 			       | Package Import 				            |
| `@Conditional` 						                 | `NewConditional()` 			         |
| `@ConditionalOnExpression` 			        | `NewExpressionCondition()` 	   |
| `@ConditionalOnProperty` 				         | `NewPropertyValueCondition()`	 |
| `@ConditionalOnBean` 					            | `NewBeanCondition()` 			       |
| `@ConditionalOnMissingBean` 			       | `NewMissingBeanCondition()`	   |
| `@ConditionalOnClass` 				            | Don't Need 					               |
| `@ConditionalOnMissingClass` 			      | Don't Need 					               |
| `@Lookup` 							                     | —— 							                     |

### Property binding

`Go-Spring` not only supports property binding for primitive data types but also supports property binding for custom value types. It also provides support for nested binding of struct properties.

```go
type DB struct {
	UserName string `value:"${username}"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
	DB       string `value:"${db}"`
}

type DbConfig struct {
	DB []DB `value:"${db}"`
}
```

The above code can be bound using the following configuration：

```yaml
db:
  -
    username: root
    password: 123456
    url: 1.1.1.1
    port: 3306
    db: db1
  -
    username: root
    password: 123456
    url: 1.1.1.1
    port: 3306
    db: db2
```

### License

The `Go-Spring` is released under version 2.0 of the Apache License.