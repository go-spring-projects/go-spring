# Go-Spring
Go-Spring 的愿景是让 Go 程序员也能用上如 Java Spring 那般威力强大的编程框架，立志为用户提供简单、安全、可信赖的编程体验。

本项目修改自： [go-spring/go-spring](https://github.com/go-spring/go-spring)
* 采用主库发布模式
* 精简第三方模块，仅保留核心依赖注入

### IoC 容器

Go-Spring 不仅实现了如 Java Spring 那般功能强大的 IoC 容器，还扩充了 Bean 的概念。在 Go 中，对象(即指针)、数组、Map、函数指针，这些都是 Bean，都可以放到 IoC 容器里。

| 常用的 Java Spring 注解				                | 对应的 Go-Spring 实现			            |
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

### 属性绑定

Go-Spring 不仅支持对普通数据类型进行属性绑定，也支持对自定义值类型进行属性绑定，而且还支持对结构体属性的嵌套绑定。

```
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

上面这段代码可以通过下面的配置进行绑定：

```
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

#### 发起者

[@lvan100 (LiangHuan)](https://github.com/lvan100)

### License

The Go-Spring is released under version 2.0 of the Apache License.