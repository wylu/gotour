# 2.2 进行项目设计

在完成了初步的示例演示后，接下来就是进入具体的预备开发阶段，一般在正式进入业务开发前，我们会针对本次需求的迭代内容进行多类的设计和评审，无设计不开发。但是问题在于，我们目前还缺很多初始化的东西没有做，因此在本章节中，我们主要针对项目目录结构、接口方案、路由注册、数据库等设计进行思考和设计开发。

## 2.2.1 目录结构

我们先将项目的标准目录结构创建起来，便于后续的开发，最终目录结构如下：

```shell
blog-service
├── configs
├── docs
├── global
├── internal
│   ├── dao
│   ├── middleware
│   ├── model
│   ├── routers
│   └── service
├── pkg
├── storage
├── scripts
└── third_party
```

- configs：配置文件。
- docs：文档集合。
- global：全局变量。
- internal：内部模块。
  - dao：数据访问层（Database Access Object），所有与数据相关的操作都会在 dao 层进行，例如 MySQL、ElasticSearch 等。
  - middleware：HTTP 中间件。
  - model：模型层，用于存放 model 对象。
  - routers：路由相关逻辑处理。
  - service：项目核心业务逻辑。
- pkg：项目相关的模块包。
- storage：项目生成的临时文件。
- scripts：各类构建，安装，分析等操作的脚本。
- third_party：第三方的资源工具，例如 Swagger UI。

## 2.2.2 数据库

在本次的项目开发中，我们主要是要实现两大块的基础业务功能，功能点分别如下：

- 标签管理：文章所归属的分类，也就是标签。我们平时都会针对文章的内容打上好几个标签，用于标识文章内容的要点要素，这样子便于读者的识别和 SEO 的收录等。
- 文章管理：整个文章内容的管理，并且需要将文章和标签进行关联。

那么要做业务开发，第一点就是要设计数据库，因此我们将根据业务模块来进行 MySQL 数据库的创建和进行表设计，概述如下：

![image](https://golang2.eddycjy.com/images/ch2/db-design.jpg)

### 2.2.2.1 创建数据库

首先你需要准备一个 MySQL 数据库，版本使用 5.7 就可以了，并在 MySQL 中执行如下 SQL 语句：

```shell
CREATE DATABASE
IF
	NOT EXISTS blog_service DEFAULT CHARACTER 
	SET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;
```

通过上述 SQL 语句，数据库将会创建本项目的数据库 blog_service，并设置它的默认编码为 utf8mb4。另外在每个数据表中，都包含同样的公共字段，如下：

```shell
  `created_on` int(10) unsigned DEFAULT '0' COMMENT '创建时间',
  `created_by` varchar(100) DEFAULT '' COMMENT '创建人',
  `modified_on` int(10) unsigned DEFAULT '0' COMMENT '修改时间',
  `modified_by` varchar(100) DEFAULT '' COMMENT '修改人',
  `deleted_on` int(10) unsigned DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint(3) unsigned DEFAULT '0' COMMENT '是否删除 0 为未删除、1 为已删除',
```

大家在创建数据表时，注意将其同时包含写入就可以了。

### 2.2.2.2 创建标签表

```shell
CREATE TABLE `blog_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) DEFAULT '' COMMENT '标签名称',
  # 此处请写入公共字段
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0 为禁用、1 为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签管理';
```

创建标签表，表字段主要为标签的名称、状态以及公共字段。

### 2.2.2.3 创建文章表

```shell
CREATE TABLE `blog_article` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(100) DEFAULT '' COMMENT '文章标题',
  `desc` varchar(255) DEFAULT '' COMMENT '文章简述',
  `cover_image_url` varchar(255) DEFAULT '' COMMENT '封面图片地址',
  `content` longtext COMMENT '文章内容',
  # 此处请写入公共字段
  `state` tinyint(3) unsigned DEFAULT '1' COMMENT '状态 0 为禁用、1 为启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章管理';
```

创建文章表，表字段主要为文章的标题、封面图、内容概述以及公共字段。

### 2.2.2.4 创建文章标签关联表

```shell
CREATE TABLE `blog_article_tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `article_id` int(11) NOT NULL COMMENT '文章 ID',
  `tag_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '标签 ID',
  # 此处请写入公共字段
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章标签关联';
```

创建文章标签关联表，这个表主要用于记录文章和标签之间的 1:N 的关联关系。

## 2.2.3 创建 model

在完成了数据库的表创建后，我们需要到项目目录下的 `internal/model` 目录创建对应的 model 对象，便于后续应用程序的使用。

### 2.2.3.1 创建公共 model

在 `internal/model` 目录下创建 model.go 文件，写入如下代码：

```go
type Model struct {
	ID         uint32 `gorm:"primary_key" json:"id"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	CreatedOn  uint32 `json:"created_on"`
	ModifiedOn uint32 `json:"modified_on"`
	DeletedOn  uint32 `json:"deleted_on"`
	IsDel      uint8  `json:"is_del"`
}
```

### 2.2.3.2 创建标签 model

在 `internal/model` 目录下创建 tag.go 文件，写入如下代码：

```go
type Tag struct {
	*Model
	Name  string `json:"name"`
	State uint8  `json:"state"`
}

func (t Tag) TableName() string {
	return "blog_tag"
}
```

### 2.2.3.3 创建文章 model

在 `internal/model` 目录下创建 article.go 文件，写入如下代码：

```go
type Article struct {
	*Model
	Title         string `json:"title"`
	Desc          string `json:"desc"`
	Content       string `json:"content"`
	CoverImageUrl string `json:"cover_image_url"`
	State         uint8  `json:"state"`
}

func (a Article) TableName() string {
	return "blog_article"
}
```

### 2.2.3.4 创建文章标签 model

在 `internal/model` 目录下创建 article_tag.go 文件，写入如下代码：

```go
type ArticleTag struct {
	*Model
	TagID     uint32 `json:"tag_id"`
	ArticleID uint32 `json:"article_id"`
}

func (a ArticleTag) TableName() string {
	return "blog_article_tag"
}
```

## 2.2.4 路由

在完成数据库的设计后，我们需要对业务模块的管理接口进行设计，而在这一块最核心的就是增删改查的 RESTful API 设计和编写，在 RESTful API 中 HTTP 方法对应的行为动作分别如下：

- GET：读取/检索动作。
- POST：新增/新建动作。
- PUT：更新动作，用于更新一个完整的资源，要求为幂等。
- PATCH：更新动作，用于更新某一个资源的一个组成部分，也就是只需要更新该资源的某一项，就应该使用 PATCH 而不是 PUT，可以不幂等。
- DELETE：删除动作。

接下来在下面的小节中，我们就可以根据 RESTful API 的基本规范，针对我们的业务模块设计路由规则，从业务角度来划分多个管理接口。

### 2.2.4.1 标签管理

| 功能         | HTTP 方法 | 路径      |
| ------------ | --------- | --------- |
| 新增标签     | POST      | /tags     |
| 删除指定标签 | DELETE    | /tags/:id |
| 更新指定标签 | PUT       | /tags/:id |
| 获取标签列表 | GET       | /tags     |

### 2.2.4.2 文章管理

| 功能         | HTTP 方法 | 路径          |
| ------------ | --------- | ------------- |
| 新增文章     | POST      | /articles     |
| 删除指定文章 | DELETE    | /articles/:id |
| 更新指定文章 | PUT       | /articles/:id |
| 获取指定文章 | GET       | /articles/:id |
| 获取文章列表 | GET       | /articles     |

### 2.2.4.3 路由管理

在确定了业务接口设计后，需要对业务接口进行一个基础编码，确定其方法原型，把当前工作区切换到项目目录的 `internal/routers` 下，并新建 router.go 文件，写入代码：

```go
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	apiv1 := r.Group("/api/v1")
	{
		apiv1.POST("/tags")
		apiv1.DELETE("/tags/:id")
		apiv1.PUT("/tags/:id")
		apiv1.PATCH("/tags/:id/state")
		apiv1.GET("/tags")
		
		apiv1.POST("/articles")
		apiv1.DELETE("/articles/:id")
		apiv1.PUT("/articles/:id")
		apiv1.PATCH("/articles/:id/state")
		apiv1.GET("/articles/:id")
		apiv1.GET("/articles")
	}

	return r
}
```

## 2.2.5 处理程序

接下来编写对应路由的处理方法，我们在项目目录下新建 `internal/routers/api/v1` 文件夹，并新建 tag.go（标签）和 article.go（文章）文件，写入代码下述代码。

### 2.2.5.1 tag.go 文件

```go
type Tag struct {}

func NewTag() Tag {
	return Tag{}
}

func (t Tag) Get(c *gin.Context) {}
func (t Tag) List(c *gin.Context) {}
func (t Tag) Create(c *gin.Context) {}
func (t Tag) Update(c *gin.Context) {}
func (t Tag) Delete(c *gin.Context) {}
```

### 2.2.5.2 article.go 文件

```go
type Article struct{}

func NewArticle() Article {
	return Article{}
}

func (a Article) Get(c *gin.Context) {}
func (a Article) List(c *gin.Context) {}
func (a Article) Create(c *gin.Context) {}
func (a Article) Update(c *gin.Context) {}
func (a Article) Delete(c *gin.Context) {}
```

### 2.2.5.3 路由管理

在编写好路由的 Handler 方法后，我们只需要将其注册到对应的路由规则上就好了，打开项目目录下 `internal/routers` 的 router.go 文件，修改如下：

```go
  ...
  article := v1.NewArticle()
  tag := v1.NewTag()
  apiv1 := r.Group("/api/v1")
  {
      apiv1.POST("/tags", tag.Create)
      apiv1.DELETE("/tags/:id", tag.Delete)
      apiv1.PUT("/tags/:id", tag.Update)
      apiv1.PATCH("/tags/:id/state", tag.Update)
      apiv1.GET("/tags", tag.List)

      apiv1.POST("/articles", article.Create)
      apiv1.DELETE("/articles/:id", article.Delete)
      apiv1.PUT("/articles/:id", article.Update)
      apiv1.PATCH("/articles/:id/state", article.Update)
      apiv1.GET("/articles/:id", article.Get)
      apiv1.GET("/articles", article.List)
  }
```

## 2.2.6 启动接入

在完成了模型、路由的代码编写后，我们修改前面章节所编写的 main.go 文件，把它改造为这个项目的启动文件，修改代码如下：

```go
func main() {
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
```

我们通过自定义 `http.Server`，设置了监听的 TCP Endpoint、处理的程序、允许读取/写入的最大时间、请求头的最大字节数等基础参数，最后调用 `ListenAndServe` 方法开始监听。

## 2.2.7 验证

我们在项目根目录下，执行 `go run main.go` 将这个服务运行起来，查看服务是否正常运行，如下：

```shell
$ go run main.go
...
[GIN-debug] POST   /api/v1/tags              --> github.com/go-programming-tour-book/blog-service/internal/routers/api/v1.Tag.Create-fm (3 handlers)
[GIN-debug] DELETE /api/v1/tags/:id          --> github.com/go-programming-tour-book/blog-service/internal/routers/api/v1.Tag.Delete-fm (3 handlers)
```

启动信息表示路由正常注册，你可以再去实际调用一下接口，看看返回是不是正常，这一节就大功告成了。

# 2.3 编写公共组件

刚想正式的开始编码，你会突然发现，怎么什么配套组件都没有，写起来一点都不顺手，没法形成闭环。

实际上在我们每个公司的项目中，都会有一类组件，我们常称其为基础组件，又或是公共组件，它们是不带强业务属性的，串联着整个应用程序，一般由负责基建或第一批搭建的该项目的同事进行梳理和编写，如果没有这类组件，谁都写一套，是非常糟糕的，并且这个应用程序是无法形成闭环的。

因此在这一章节我们将完成一个 Web 应用中最常用到的一些基础组件，保证应用程序的标准化，一共分为如下五个板块：

![image](https://golang2.eddycjy.com/images/ch2/common-component.jpg)

## 2.3.1 错误码标准化

在应用程序的运行中，我们常常需要与客户端进行交互，而交互分别是两点，一个是正确响应下的结果集返回，另外一个是错误响应的错误码和消息体返回，用于告诉客户端，这一次请求发生了什么事，因为什么原因失败了。而在错误码的处理上，又延伸出一个新的问题，那就是错误码的标准化处理，不提前预判，将会造成比较大的麻烦，如下：

![image](https://golang2.eddycjy.com/images/ch2/errcode.jpg)

在上图中，我们可以看到客户端分别调用了三个不同的服务端，三个服务端 A、B、C，它们的响应结果的模式都不一样…如果不做任何挣扎的话，那客户端就需要知道它调用的是哪个服务，然后每一个服务写一种错误码处理规则，非常麻烦，那如果后面继续添加新的服务端，如果又不一样，那岂不是适配的更加多了？

至少在大的层面来讲，我们要尽可能的保证每个项目前后端的交互语言规则是一致的，因此在一个新项目搭建之初，其中重要的一项预备工作，那就是标准化我们的错误码格式，保证客户端是“理解”我们的错误码规则，不需要每次都写一套新的。

### 2.3.1.1 公共错误码

我们需要在在项目目录下的 `pkg/errcode` 目录新建 common_code.go 文件，用于预定义项目中的一些公共错误码，便于引导和规范大家的使用，如下：

```go
var (
	Success                   = NewError(0, "成功")
	ServerError               = NewError(10000000, "服务内部错误")
	InvalidParams             = NewError(10000001, "入参错误")
	NotFound                  = NewError(10000002, "找不到")
	UnauthorizedAuthNotExist  = NewError(10000003, "鉴权失败，找不到对应的 AppKey 和 AppSecret")
	UnauthorizedTokenError    = NewError(10000004, "鉴权失败，Token 错误")
	UnauthorizedTokenTimeout  = NewError(10000005, "鉴权失败，Token 超时")
	UnauthorizedTokenGenerate = NewError(10000006, "鉴权失败，Token 生成失败")
	TooManyRequests           = NewError(10000007, "请求过多")
)
```

### 2.3.1.2 错误处理

接下来我们在项目目录下的 `pkg/errcode` 目录新建 errcode.go 文件，编写常用的一些错误处理公共方法，标准化我们的错误输出，如下：

```go
type Error struct {
	code int `json:"code"`
	msg string `json:"msg"`
	details []string `json:"details"`
}

var codes = map[int]string{}

func NewError(code int, msg string) *Error {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已经存在，请更换一个", code))
	}
	codes[code] = msg
	return &Error{code: code, msg: msg}
}

func (e *Error) Error() string {
	return fmt.Sprintf("错误码：%d, 错误信息:：%s", e.Code(), e.Msg())
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) Msgf(args []interface{}) string {
	return fmt.Sprintf(e.msg, args...)
}

func (e *Error) Details() []string {
	return e.details
}

func (e *Error) WithDetails(details ...string) *Error {
	newError := *e
	newError.details = []string{}
	for _, d := range details {
		newError.details = append(newError.details, d)
	}

	return &newError
}

func (e *Error) StatusCode() int {
	switch e.Code() {
	case Success.Code():
		return http.StatusOK
	case ServerError.Code():
		return http.StatusInternalServerError
	case InvalidParams.Code():
		return http.StatusBadRequest
	case UnauthorizedAuthNotExist.Code():
		fallthrough
	case UnauthorizedTokenError.Code():
		fallthrough
	case UnauthorizedTokenGenerate.Code():
		fallthrough
	case UnauthorizedTokenTimeout.Code():
		return http.StatusUnauthorized
	case TooManyRequests.Code():
		return http.StatusTooManyRequests
	}

	return http.StatusInternalServerError
}
```

在错误码方法的编写中，我们声明了 `Error` 结构体用于表示错误的响应结果，并利用 `codes` 作为全局错误码的存储载体，便于查看当前注册情况，并在调用 `NewError` 创建新的 `Error` 实例的同时进行排重的校验。

另外相对特殊的是 `StatusCode` 方法，它主要用于针对一些特定错误码进行状态码的转换，因为不同的内部错误码在 HTTP 状态码中都代表着不同的意义，我们需要将其区分开来，便于客户端以及监控/报警等系统的识别和监听。

## 2.3.2 配置管理

在应用程序的运行生命周期中，最直接的关系之一就是应用的配置读取和更新。它的一举一动都有可能影响应用程序的改变，其分别包含如下行为：

![image](https://golang2.eddycjy.com/images/ch2/config.jpg)

- 在启动时：可以进行一些基础应用属性、连接第三方实例（MySQL、NoSQL）等等的初始化行为。
- 在运行中：可以监听文件或其他存储载体的变更来实现热更新配置的效果，例如：在发现有变更的话，就对原有配置值进行修改，以此达到相关联的一个效果。如果更深入业务使用的话，我们还可以通过配置的热更新，达到功能灰度的效果，这也是一个比较常见的场景。

另外，配置组件是会根据实际情况去选型的，一般大多为文件配置或配置中心的模式，在本次博客后端中我们的配置管理使用最常见的文件配置作为我们的选型。

### 2.3.2.1 安装

为了完成文件配置的读取，我们需要借助第三方开源库 viper，在项目根目录下执行以下安装命令：

```shell
$ go get -u github.com/spf13/viper@v1.4.0
```

Viper 是适用于 Go 应用程序的完整配置解决方案，是目前 Go 语言中比较流行的文件配置解决方案，它支持处理各种不同类型的配置需求和配置格式。

### 2.3.2.2 配置文件

在项目目录下的 `configs` 目录新建 config.yaml 文件，写入以下配置：

```shell
Server:
  RunMode: debug
  HttpPort: 8000
  ReadTimeout: 60
  WriteTimeout: 60
App:
  DefaultPageSize: 10
  MaxPageSize: 100
  LogSavePath: storage/logs
  LogFileName: app
  LogFileExt: .log
Database:
  DBType: mysql
  Username: root  # 填写你的数据库账号
  Password: rootroot  # 填写你的数据库密码
  Host: 127.0.0.1:3306
  DBName: blog_service
  TablePrefix: blog_
  Charset: utf8
  ParseTime: True
  MaxIdleConns: 10
  MaxOpenConns: 30
```

在配置文件中，我们分别针对如下内容进行了默认配置：

- Server：服务配置，设置 gin 的运行模式、默认的 HTTP 监听端口、允许读取和写入的最大持续时间。
- App：应用配置，设置默认每页数量、所允许的最大每页数量以及默认的应用日志存储路径。
- Database：数据库配置，主要是连接实例所必需的基础参数。

### 2.3.2.3 编写组件

在完成了配置文件的确定和编写后，我们需要针对读取配置的行为进行封装，便于应用程序的使用，我们在项目目录下的 `pkg/setting` 目录下新建 setting.go 文件，写入如下代码：

```go
type Setting struct {
	vp *viper.Viper
}

func NewSetting() (*Setting, error) {
	vp := viper.New()
	vp.SetConfigName("config")
	vp.AddConfigPath("configs/")
	vp.SetConfigType("yaml")
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return &Setting{vp}, nil
}
```

在这里我们编写了 `NewSetting` 方法，用于初始化本项目的配置的基础属性，设定配置文件的名称为 `config`，配置类型为 `yaml`，并且设置其配置路径为相对路径 `configs/`，以此确保在项目目录下执行运行时能够成功启动。

另外 viper 是允许设置多个配置路径的，这样子可以尽可能的尝试解决路径查找的问题，也就是可以不断地调用 `AddConfigPath` 方法，这块在后续会再深入介绍。

接下来我们新建 section.go 文件，用于声明配置属性的结构体并编写读取区段配置的配置方法，如下：

```go
type ServerSettingS struct {
	RunMode      string
	HttpPort     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AppSettingS struct {
	DefaultPageSize int
	MaxPageSize     int
	LogSavePath     string
	LogFileName     string
	LogFileExt      string
}

type DatabaseSettingS struct {
	DBType       string
	UserName     string
	Password     string
	Host         string
	DBName       string
	TablePrefix  string
	Charset      string
	ParseTime    bool
	MaxIdleConns int
	MaxOpenConns int
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}

	return nil
}
```

### 2.3.2.4 包全局变量

在读取了文件的配置信息后，还是不够的，因为我们需要将配置信息和应用程序关联起来，我们才能够去使用它，因此在项目目录下的 `global` 目录下新建 setting.go 文件，写入如下代码：

```go
var (
	ServerSetting   *setting.ServerSettingS
	AppSetting      *setting.AppSettingS
	DatabaseSetting *setting.DatabaseSettingS
)
```

我们针对最初预估的三个区段配置，进行了全局变量的声明，便于在接下来的步骤将其关联起来，并且提供给应用程序内部调用。

另外全局变量的初始化，是会随着应用程序的不断演进不断改变的，因此并不是一成不变，也就是这里展示的并不一定是最终的结果。

### 2.3.2.5 初始化配置读取

在完成了所有的预备行为后，我们回到项目根目录下的 main.go 文件，修改代码如下：

```go
func init() {
	err := setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
}

func main() {...}

func setupSetting() error {
	setting, err := setting.NewSetting()
	if err != nil {
		return err
	}
	err = setting.ReadSection("Server", &global.ServerSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("App", &global.AppSetting)
	if err != nil {
		return err
	}
	err = setting.ReadSection("Database", &global.DatabaseSetting)
	if err != nil {
		return err
	}

	global.ServerSetting.ReadTimeout *= time.Second
	global.ServerSetting.WriteTimeout *= time.Second
	return nil
}
```

我们新增了一个 `init` 方法，有的读者可能会疑惑它有什么作用，在 Go 语言中，`init` 方法常用于应用程序内的一些初始化操作，它在 `main` 方法之前自动执行，它的执行顺序是：全局变量初始化 =》init 方法 =》main 方法，但并不是建议滥用，因为如果 `init` 过多，你可能会迷失在各个库的 `init` 方法中，会非常麻烦。

而在我们的应用程序中，该 `init` 方法主要作用是进行应用程序的初始化流程控制，整个应用代码里也只会有一个 `init` 方法，因此我们在这里调用了初始化配置的方法，达到配置文件内容映射到应用配置结构体的作用。

### 2.3.2.7 修改服务端配置

接下来我们只需要在启动文件 main.go 中把已经映射好的配置和 gin 的运行模式进行设置，这样的话，在程序重新启动时后就可以生效，如下：

```go
func main() {
	gin.SetMode(global.ServerSetting.RunMode)
	router := routers.NewRouter()
	s := &http.Server{
		Addr:           ":" + global.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    global.ServerSetting.ReadTimeout,
		WriteTimeout:   global.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
```

### 2.3.2.8 验证

在完成了配置相关的初始化后，我们需要校验配置是否真正的映射到配置结构体上了，我们一般可以通过断点或简单打日志的方式进行查看，最终配置的包全局变量的值应当要得出如下结果：

```shell
global.ServerSetting: &{RunMode:debug HttpPort:8000 ReadTimeout:1m0s WriteTimeout:1m0s}

global.AppSetting: &{DefaultPageSize:10 MaxPageSize:100}

global.DatabaseSetting: &{DBType:mysql User: Password:rootroot Host:127.0.0.1:3306 DBName:blog TablePrefix:blog_}
```

## 2.3.3 数据库连接

### 2.3.3.1 安装

我们在本项目中数据库相关的数据操作将使用第三方的开源库 gorm，它是目前 Go 语言中最流行的 ORM 库（从 Github Star 来看），同时它也是一个功能齐全且对开发人员友好的 ORM 库，目前在 Github 上相当的活跃，具有一定的保障，安装命令如下：

```shell
$ go get -u github.com/jinzhu/gorm@v1.9.12
```

另外在社区中，也有其它的声音，例如有认为不使用 ORM 库更好的，这类的比较本文暂不探讨，但若是想了解的话可以看看像 sqlx 这类 database/sql 的扩展库，也是一个不错的选择。

### 2.3.3.2 编写组件

我们打开项目目录 `internal/model` 下的 model.go 文件，新增 NewDBEngine 方法，如下：

```go
import (
	...
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Model struct {...}

func NewDBEngine(databaseSetting *setting.DatabaseSettingS) (*gorm.DB, error) {
	db, err := gorm.Open(databaseSetting.DBType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local",
		databaseSetting.UserName,
		databaseSetting.Password,
		databaseSetting.Host,
		databaseSetting.DBName,
		databaseSetting.Charset,
		databaseSetting.ParseTime,
	))
	if err != nil {
		return nil, err
	}
	
	if global.ServerSetting.RunMode == "debug" {
		db.LogMode(true)
	}
	db.SingularTable(true)
	db.DB().SetMaxIdleConns(databaseSetting.MaxIdleConns)
	db.DB().SetMaxOpenConns(databaseSetting.MaxOpenConns)

	return db, nil
}
```

我们通过上述代码，编写了一个针对创建 DB 实例的 NewDBEngine 方法，同时增加了 gorm 开源库的引入和 MySQL 驱动库 `github.com/jinzhu/gorm/dialects/mysql` 的初始化（不同类型的 DBType 需要引入不同的驱动库，否则会存在问题）。

### 2.3.3.3 包全局变量

我们在项目目录下的 `global` 目录，新增 db.go 文件，新增如下内容：

```go
var (
	DBEngine *gorm.DB
)
```

### 2.3.3.4 初始化

回到启动文件，也就是项目目录下的 main.go 文件，新增 setupDBEngine 方法初始化，如下：

```go
func init() {
	...
	err = setupDBEngine()
	if err != nil {
		log.Fatalf("init.setupDBEngine err: %v", err)
	}
}

func main() {...}
func setupSetting() error {...}
func setupLogger() error {...}

func setupDBEngine() error {
	var err error
	global.DBEngine, err = model.NewDBEngine(global.DatabaseSetting)
	if err != nil {
		return err
	}

	return nil
}
```

这里需要注意，有一些人会把初始化语句不小心写成：`global.DBEngine, err := model.NewDBEngine(global.DatabaseSetting)`，这是存在很大问题的，因为 `:=` 会重新声明并创建了左侧的新局部变量，因此在其它包中调用 `global.DBEngine` 变量时，它仍然是 `nil`，仍然是达不到可用标准，因为根本就没有赋值到真正需要赋值的包全局变量 `global.DBEngine` 上。

## 2.3.4 日志写入

如果有心的读者会发现我们在上述应用代码中都是直接使用 Go 标准库 log 来进行的日志输出，这其实是有些问题的，因为在一个项目中，我们的日志需要标准化的记录一些的公共信息，例如：代码调用堆栈、请求链路 ID、公共的业务属性字段等等，而直接输出标准库的日志的话，并不具备这些数据，也不够灵活。

日志的信息的齐全与否在排查和调试问题中是非常重要的一环，因此在应用程序中我们也会有一个标准的日志组件会进行统一处理和输出。

### 2.3.4.1 安装

```shell
$ go get -u gopkg.in/natefinch/lumberjack.v2
```

我们先拉取日志组件内要使用到的第三方的开源库 lumberjack，它的核心功能是将日志写入滚动文件中，该库支持设置所允许单日志文件的最大占用空间、最大生存周期、允许保留的最多旧文件数，如果出现超出设置项的情况，就会对日志文件进行滚动处理。

而我们使用这个库，主要是为了减免一些文件操作类的代码编写，把核心逻辑摆在日志标准化处理上。

### 2.3.4.2 编写组件

首先在这一节中，实质上代码都是在同一个文件中的，但是为了便于理解，我们会在讲解上会将日志组件的代码切割为多块进行剖析。

#### 2.3.4.2.1 日志分级

我们在项目目录下的 `pkg/` 目录新建 `logger` 目录，并创建 logger.go 文件，写入日志分级相关的代码：

```go
type Level int8

type Fields map[string]interface{}

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	}
	return ""
}
```

我们先预定义了应用日志的 Level 和 Fields 的具体类型，并且分为了 Debug、Info、Warn、Error、Fatal、Panic 六个日志等级，便于在不同的使用场景中记录不同级别的日志。

#### 2.3.4.2.2 日志标准化

我们完成了日志的分级方法后，开始编写具体的方法去进行日志的实例初始化和标准化参数绑定，继续写入如下代码：

```go
type Logger struct {
	newLogger *log.Logger
	ctx       context.Context
	fields    Fields
	callers   []string
}

func NewLogger(w io.Writer, prefix string, flag int) *Logger {
	l := log.New(w, prefix, flag)
	return &Logger{newLogger: l}
}

func (l *Logger) clone() *Logger {
	nl := *l
	return &nl
}

func (l *Logger) WithFields(f Fields) *Logger {
	ll := l.clone()
	if ll.fields == nil {
		ll.fields = make(Fields)
	}
	for k, v := range f {
		ll.fields[k] = v
	}
	return ll
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	ll := l.clone()
	ll.ctx = ctx
	return ll
}

func (l *Logger) WithCaller(skip int) *Logger {
	ll := l.clone()
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		f := runtime.FuncForPC(pc)
		ll.callers = []string{fmt.Sprintf("%s: %d %s", file, line, f.Name())}
	}

	return ll
}

func (l *Logger) WithCallersFrames() *Logger {
	maxCallerDepth := 25
	minCallerDepth := 1
	callers := []string{}
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		callers = append(callers, fmt.Sprintf("%s: %d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	ll := l.clone()
	ll.callers = callers
	return ll
}
```

- WithLevel：设置日志等级。
- WithFields：设置日志公共字段。
- WithContext：设置日志上下文属性。
- WithCaller：设置当前某一层调用栈的信息（程序计数器、文件信息、行号）。
- WithCallersFrames：设置当前的整个调用栈信息。

#### 2.3.4.2.3 日志格式化和输出

我们开始编写日志内容的格式化和日志输出动作的相关方法，继续写入如下代码：

```go
func (l *Logger) JSONFormat(level Level, message string) map[string]interface{} {
	data := make(Fields, len(l.fields)+4)
	data["level"] = level.String()
	data["time"] = time.Now().Local().UnixNano()
	data["message"] = message
	data["callers"] = l.callers
	if len(l.fields) > 0 {
		for k, v := range l.fields {
			if _, ok := data[k]; !ok {
				data[k] = v
			}
		}
	}

	return data
}

func (l *Logger) Output(level Level, message string) {
	body, _ := json.Marshal(l.JSONFormat(level, message))
	content := string(body)
	switch level {
	case LevelDebug:
		l.newLogger.Print(content)
	case LevelInfo:
		l.newLogger.Print(content)
	case LevelWarn:
		l.newLogger.Print(content)
	case LevelError:
		l.newLogger.Print(content)
	case LevelFatal:
		l.newLogger.Fatal(content)
	case LevelPanic:
		l.newLogger.Panic(content)
	}
}
```

#### 2.3.4.2.4 日志分级输出

我们根据先前定义的日志分级，编写对应的日志输出的外部方法，继续写入如下代码：

```go
func (l *Logger) Info(v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprint(v...))
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(LevelFatal, fmt.Sprintf(format, v...))
}
...
```

上述代码中仅展示了 Info、Fatal 级别的日志方法，这里主要是根据 Debug、Info、Warn、Error、Fatal、Panic 六个日志等级编写对应的方法，大家可自行完善，除了方法名以及 WithLevel 设置的不一样，其他均为一致的代码。

### 2.3.4.3 包全局变量

在完成日志库的编写后，我们需要定义一个 Logger 对象便于我们的应用程序使用。因此我们打开项目目录下的 `global/setting.go` 文件，新增如下内容：

```go
var (
	...
	Logger          *logger.Logger
)
```

我们在包全局变量中新增了 Logger 对象，用于日志组件的初始化。

### 2.3.4.4 初始化

接下来我们需要修改启动文件，也就是项目目录下的 main.go 文件，新增对刚刚定义的 Logger 对象的初始化，如下：

```go
func init() {
	err := setupSetting()
	if err != nil {
		log.Fatalf("init.setupSetting err: %v", err)
	}
	err = setupLogger()
	if err != nil {
		log.Fatalf("init.setupLogger err: %v", err)
	}
}

func main() {...}
func setupSetting() error {...}

func setupLogger() error {
	global.Logger = logger.NewLogger(&lumberjack.Logger{
		Filename: global.AppSetting.LogSavePath + "/" + global.AppSetting.LogFileName + global.AppSetting.LogFileExt,
		MaxSize:   600,
		MaxAge:    10,
		LocalTime: true,
	}, "", log.LstdFlags).WithCaller(2)

	return nil
}
```

通过这段程序，我们在 init 方法中新增了日志组件的流程，并在 setupLogger 方法内部对 global 的包全局变量 Logger 进行了初始化，需要注意的是我们使用了 lumberjack 作为日志库的 io.Writer，并且设置日志文件所允许的最大占用空间为 600MB、日志文件最大生存周期为 10 天，并且设置日志文件名的时间格式为本地时间。

### 2.3.4.5 验证

在完成了上述的步骤后，日志组件已经正式的初始化完毕了，为了验证你是否操作正确，你可以在 main 方法中执行下述测试代码：

```shell
global.Logger.Infof("%s: go-programming-tour-book/%s", "eddycjy", "blog-service")
```

接着可以查看项目目录下的 `storage/logs/app.log`，看看日志文件是否正常创建且写入了预期的日志记录，大致如下：

```shell
{"callers":["~/go-programming-tour-book/blog-service/main.go: 20 main.init.0"],"level":"info","message":"eddycjy: go-programming-tour-book/blog-service","time":xxxx}
```

## 2.3.5 响应处理

在应用程序中，与客户端对接的常常是服务端的接口，那客户端是怎么知道这一次的接口调用结果是怎么样的呢？一般来讲，主要是通过对返回的 HTTP 状态码和接口返回的响应结果进行判断，而判断的依据则是事先按规范定义好的响应结果。

因此在这一小节，我们将编写统一处理接口返回的响应处理方法，它也正正与错误码标准化是相对应的。

### 2.3.5.1 类型转换

在项目目录下的 `pkg/convert` 目录下新建 convert.go 文件，如下：

```go
type StrTo string

func (s StrTo) String() string {
	return string(s)
}

func (s StrTo) Int() (int, error) {
	v, err := strconv.Atoi(s.String())
	return v, err
}

func (s StrTo) MustInt() int {
	v, _ := s.Int()
	return v
}

func (s StrTo) UInt32() (uint32, error) {
	v, err := strconv.Atoi(s.String())
	return uint32(v), err
}

func (s StrTo) MustUInt32() uint32 {
	v, _ := s.UInt32()
	return v
}
```

### 2.3.5.2 分页处理

在项目目录下的 `pkg/app` 目录下新建 pagination.go 文件，如下：

```go
func GetPage(c *gin.Context) int {
	page := convert.StrTo(c.Query("page")).MustInt()
	if page <= 0 {
		return 1
	}

	return page
}

func GetPageSize(c *gin.Context) int {
	pageSize := convert.StrTo(c.Query("page_size")).MustInt()
	if pageSize <= 0 {
		return global.AppSetting.DefaultPageSize
	}
	if pageSize > global.AppSetting.MaxPageSize {
		return global.AppSetting.MaxPageSize
	}

	return pageSize
}

func GetPageOffset(page, pageSize int) int {
	result := 0
	if page > 0 {
		result = (page - 1) * pageSize
	}

	return result
}
```

### 2.3.5.3 响应处理

在项目目录下的 `pkg/app` 目录下新建 app.go 文件，如下：

```go
type Response struct {
	Ctx *gin.Context
}

type Pager struct {
	Page int `json:"page"`
	PageSize int `json:"page_size"`
	TotalRows int `json:"total_rows"`
}

func NewResponse(ctx *gin.Context) *Response {
	return &Response{Ctx: ctx}
}

func (r *Response) ToResponse(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	r.Ctx.JSON(http.StatusOK, data)
}

func (r *Response) ToResponseList(list interface{}, totalRows int) {
	r.Ctx.JSON(http.StatusOK, gin.H{
		"list": list,
		"pager": Pager{
			Page:      GetPage(r.Ctx),
			PageSize:  GetPageSize(r.Ctx),
			TotalRows: totalRows,
		},
	})
}

func (r *Response) ToErrorResponse(err *errcode.Error) {
	response := gin.H{"code": err.Code(), "msg": err.Msg()}
	details := err.Details()
	if len(details) > 0 {
		response["details"] = details
	}

	r.Ctx.JSON(err.StatusCode(), response)
}
```

### 2.3.5.4 验证

我们可以找到其中一个接口方法，调用对应的方法，检查是否有误，如下：

```go
func (a Article) Get(c *gin.Context) {
	app.NewResponse(c).ToErrorResponse(errcode.ServerError)
	return
}
```

验证响应结果，如下：

```shell
$ curl -v http://127.0.0.1:8080/api/v1/articles/1
...
< HTTP/1.1 500 Internal Server Error
{"code":10000000,"msg":"服务内部错误"}
```

从响应结果上看，可以知道本次接口的调用结果的 HTTP 状态码为 500，响应消息体为约定的错误体，符合我们的要求。

## 2.3.6 小结

在本章节中，我们主要是针对项目的公共组件初始化，做了大量的规范制定、公共库编写、初始化注册等等行为，虽然比较繁琐，这这些公共组件在整个项目运行中至关重要，早期做的越标准化，后期越省心省事，因为大家直接使用就可以了，不需要过多的关心细节，也不会有人重新再造新的公共库轮子，导致要适配多套。

# 2.4 生成接口文档

我们在前面的章节中完成了针对业务需求的模块和路由的设计，并且完成了公共组件的处理，初步运行也没有问题，那么这一次是不是真的就可以开始编码了呢？

其实不然，虽然我们完成了路由的设计，但是接口的定义不是一个人的事，我们在提前设计好接口的入参、出参以及异常情况后，还需要其他同事一起进行接口设计评审，以便确认本次迭代的接口设计方案是尽可能正确和共同认可的，如下图：

![image](https://golang2.eddycjy.com/images/ch2/api-design.jpg)

## 2.4.1 什么是 Swagger

那如何维护接口文档，是绝大部分开发人员都经历过的问题，因为前端、后端、测试开发等等人员都要看，每个人都给一份的话，怎么维护，这将是一个非常头大的问题。在很多年以前，也流行过用 Word 等等工具写接口文档，显然，这会有许许多多的问题，后端人员所耗费的精力、文档的时效性根本无法得到保障。

针对这类问题，市面上出现了大量的解决方案，Swagger 正是其中的佼佼者，它更加的全面和完善，具有相关联的生态圈。它是基于标准的 OpenAPI 规范进行设计的，只要照着这套规范去编写你的注解或通过扫描代码去生成注解，就能生成统一标准的接口文档和一系列 Swagger 工具。

## 2.4.2 OpenAPI & Swagger

在上文我们有提到 OpenAPI，你可能会对此产生疑惑，OpenAPI 和 Swagger 又是什么关系？

其实 OpenAPI 规范是在 2015 年由 OpenAPI Initiative 捐赠给 Linux 基金会的，并且 Swagger 对此更进一步的针对 OpenAPI 规范提供了大量与之相匹配的工具集，能够充分利用 OpenAPI 规范去映射生成所有与之关联的资源和操作去查看和调用 RESTful 接口，因此我们也常说 Swagger 不仅是一个“规范”，更是一个框架。

从功能使用上来讲，OpenAPI 规范能够帮助我们描述一个 API 的基本信息，比如：

- 有关该 API 的描述。
- 可用路径（/资源）。
- 在每个路径上的可用操作（获取/提交…）。
- 每个操作的输入/输出格式。

## 2.4.3 安装 Swagger

Swagger 相关的工具集会根据 OpenAPI 规范去生成各式各类的与接口相关联的内容，常见的流程是编写注解 =》调用生成库-》生成标准描述文件 =》生成/导入到对应的 Swagger 工具。

因此接下来第一步，我们要先安装 Go 对应的开源 Swagger 相关联的库，在项目 blog-service 根目录下执行安装命令，如下：

```shell
$ go get -u github.com/swaggo/swag/cmd/swag@v1.6.5
$ go get -u github.com/swaggo/gin-swagger@v1.2.0 
$ go get -u github.com/swaggo/files
$ go get -u github.com/alecthomas/template
```

验证是否安装成功，如下：

```shell
$ swag -v
swag version v1.6.5
```

如果命令行提示寻找不到 swag 文件，可以检查一下对应的 bin 目录是否已经加入到环境变量 PATH 中。

## 2.4.4 写入注解

在完成了 Swagger 关联库的安装后，我们需要针对项目里的 API 接口进行注解的编写，以便于后续在进行生成时能够正确的运行，接下来我们将使用到如下注解：

| 注解     | 描述                                                         |
| -------- | ------------------------------------------------------------ |
| @Summary | 摘要                                                         |
| @Produce | API 可以产生的 MIME 类型的列表，MIME 类型你可以简单的理解为响应类型，例如：json、xml、html 等等 |
| @Param   | 参数格式，从左到右分别为：参数名、入参类型、数据类型、是否必填、注释 |
| @Success | 响应成功，从左到右分别为：状态码、参数类型、数据类型、注释   |
| @Failure | 响应失败，从左到右分别为：状态码、参数类型、数据类型、注释   |
| @Router  | 路由，从左到右分别为：路由地址，HTTP 方法                    |

### 2.4.4.1 API

我们切换到项目目录下的 `internal/routers/api/v1` 目录，打开 tag.go 文件，写入如下注解：

```go
// @Summary 获取多个标签
// @Produce  json
// @Param name query string false "标签名称" maxlength(100)
// @Param state query int false "状态" Enums(0, 1) default(1)
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} model.Tag "成功"
// @Failure 400 {object} errcode.Error "请求错误"
// @Failure 500 {object} errcode.Error "内部错误"
// @Router /api/v1/tags [get]
func (t Tag) List(c *gin.Context) {}

// @Summary 新增标签
// @Produce  json
// @Param name body string true "标签名称" minlength(3) maxlength(100)
// @Param state body int false "状态" Enums(0, 1) default(1)
// @Param created_by body string true "创建者" minlength(3) maxlength(100)
// @Success 200 {object} model.Tag "成功"
// @Failure 400 {object} errcode.Error "请求错误"
// @Failure 500 {object} errcode.Error "内部错误"
// @Router /api/v1/tags [post]
func (t Tag) Create(c *gin.Context) {}

// @Summary 更新标签
// @Produce  json
// @Param id path int true "标签 ID"
// @Param name body string false "标签名称" minlength(3) maxlength(100)
// @Param state body int false "状态" Enums(0, 1) default(1)
// @Param modified_by body string true "修改者" minlength(3) maxlength(100)
// @Success 200 {array} model.Tag "成功"
// @Failure 400 {object} errcode.Error "请求错误"
// @Failure 500 {object} errcode.Error "内部错误"
// @Router /api/v1/tags/{id} [put]
func (t Tag) Update(c *gin.Context) {}

// @Summary 删除标签
// @Produce  json
// @Param id path int true "标签 ID"
// @Success 200 {string} string "成功"
// @Failure 400 {object} errcode.Error "请求错误"
// @Failure 500 {object} errcode.Error "内部错误"
// @Router /api/v1/tags/{id} [delete]
func (t Tag) Delete(c *gin.Context) {}
```

在这里我们只展示了标签模块的接口注解编写，接下来你应当按照注解的含义和参考上述接口注解，完成文章模块接口注解的编写。

### 2.4.4.2 Main

那么接口方法本身有了注解，那针对这个项目，能不能写注解呢，万一有很多个项目，怎么知道它是谁？实际上是可以识别出来的，我们只要针对 main 方法写入如下注解：

```go
// @title 博客系统
// @version 1.0
// @description Go 语言编程之旅：一起用 Go 做项目
// @termsOfService https://github.com/go-programming-tour-book
func main() {
	...
}
```

## 2.4.5 生成

在完成了所有的注解编写后，我们回到项目根目录下，执行如下命令：

```shell
$ swag init
```

在执行命令完毕后，会发现在 docs 文件夹生成 docs.go、swagger.json、swagger.yaml 三个文件。

## 2.4.6 路由

那注解编写完，也通过 swag init 把 Swagger API 所需要的文件都生成了，那接下来我们怎么访问接口文档呢？其实很简单，我们只需要在 routers 中进行默认初始化和注册对应的路由就可以了，打开项目目录下的 `internal/routers` 目录中的 router.go 文件，新增代码如下：

```go
import (
	...
	_ "github.com/go-programming-tour-book/blog-service/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	...
	return r
}
```

从表面上来看，主要做了两件事，分别是初始化 docs 包和注册一个针对 swagger 的路由，而在初始化 docs 包后，其 swagger.json 将会默认指向当前应用所启动的域名下的 swagger/doc.json 路径，如果有额外需求，可进行手动指定，如下：

```go
  url := ginSwagger.URL("http://127.0.0.1:8000/swagger/doc.json")
  r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
```

## 2.4.7 查看接口文档

![image](https://golang2.eddycjy.com/images/ch2/api-doc.jpg)

在完成了上述的设置后，我们重新启动服务端，在浏览器中访问 Swagger 的地址 `http://127.0.0.1:8000/swagger/index.html`，就可以看到上述图片中的 Swagger 文档展示，其主要分为三个部分，分别是项目主体信息、接口路由信息、模型信息，这三部分共同组成了我们主体内容。

## 2.4.8 发生了什么

可能会疑惑，我明明只是初始化了个 docs 包并注册了一个 Swagger 相关路由，Swagger 的文档是怎么关联上的呢，我在接口上写的注解又到哪里去了？

其实主体是与我们在章节 2.4.4 生成的文件有关的，分别是：

```shell
docs
├── docs.go
├── swagger.json
└── swagger.yaml
```

### 2.4.8.1 初始化 docs

在第一步中，我们初始化了 docs 包，对应的其实就是 docs.go 文件，因为目录下仅有一个 go 源文件，其源码如下：

```go
var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://github.com/go-programming-tour-book",
        "version": "{{.Version}}"
    },
    ...
}`

var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Title:       "博客系统",
	Description: "Go 语言编程之旅：一起用 Go 做项目",
	...
}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)
	t, _ := template.New("swagger_info").Funcs(template.FuncMap{...}).Parse(doc)
	
	var tpl bytes.Buffer
	_ = t.Execute(&tpl, sInfo)
	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
```

通过对源码的分析，我们可以得知实质上在初始化 docs 包时，会默认执行 init 方法，而在 init 方法中，会注册相关方法，主体逻辑是 swag 会在生成时去检索项目下的注解信息，然后将项目信息和接口路由信息按规范生成到包全局变量 doc 中去。

紧接着会在 ReadDoc 方法中做一些 template 的模板映射等工作，完善 doc 的输出。

### 2.4.8.2 注册路由

在上一步中，我们知道了生成的注解数据源在哪，但是它们两者又是怎么关联起来的呢，实际上与我们调用的 `ginSwagger.WrapHandler(swaggerFiles.Handler)` 有关，如下：

```go
func WrapHandler(h *webdav.Handler, confs ...func(c *Config)) gin.HandlerFunc {
	defaultConfig := &Config{URL: "doc.json"}
	...
	return CustomWrapHandler(defaultConfig, h)
}
```

实际上在调用 WrapHandler 后，swag 内部会将其默认调用的 URL 设置为 doc.json，但你可能会纠结，明明我们生成的文件里没有 doc.json，这又是从哪里来的，我们接着往下看，如下：

```go
func CustomWrapHandler(config *Config, h *webdav.Handler) gin.HandlerFunc {
	  ...
		switch path {
		case "index.html":
			index.Execute(c.Writer, &swaggerUIBundle{
				URL: config.URL,
			})
		case "doc.json":
			doc, err := swag.ReadDoc()
			if err != nil {
				panic(err)
			}
			c.Writer.Write([]byte(doc))
			return
		default:
			h.ServeHTTP(c.Writer, c.Request)
		}
	}
}
```

在 CustomWrapHandler 方法中，我们可以发现一处比较经典 switch case 的逻辑。

在第一个 case 中，处理是的 index.html，这又是什么呢，其实你可以回顾一下，我们在先前是通过 `http://127.0.0.1:8000/swagger/index.html` 访问到 Swagger 文档的，对应的便是这里的逻辑。

在第二个 case 中，就可以大致解释我们所关注的 doc.json 到底是什么，它相当于一个内部标识，会去读取我们所生成的 Swagger 注解，你也可以发现我们先前在访问的 Swagger 文档时，它顶部的文本框中 Explore 默认的就是 doc.json（也可以填写外部地址，只要输出的是对应的 Swagger 注解）。

## 2.4.9 问题

细心的读者可能会发现，我们先前在公共组件的章节已经定义好了一些基本类型的 Response 返回值，但我们在本章节编写成功响应时，是直接调用 model 作为其数据类型，如下：

```shell
// @Success 200 {object} model.Tag "成功"
```

这样写的话，就会有一个问题，如果有 model.Tag 以外的字段，例如分页，那就无法展示了。更接近实践来讲，大家在编码中常常会遇到某个对象内中的某一个字段是 interface，这个字段的类型它是不定的，也就是公共结构体，那注解又应该怎么写呢，如下情况：

```go
type Test struct {
	UserName string
	Content  interface{}
}
```

可能会有的人会忽略它，采取口头说明，但这显然是不完备的。而 swag 目前在 v1.6.3 也没有特别好的新注解方式，官方在 issue 里也曾表示过通过注解来解决这个问题是不合理的，那我们要怎么做呢？

实际上，官方给出的建议很简单，就是定义一个针对 Swagger 的对象，专门用于 Swagger 接口文档展示，我们在 `internal/model` 的 tag.go 和 article.go 文件中，新增如下代码：

```go
// tag.go
type TagSwagger struct {
	List  []*Tag
	Pager *app.Pager
}

// article.go
type ArticleSwagger struct {
	List  []*Article
	Pager *app.Pager
}
```

我们修改接口方法中对应的注解信息，如下：

```shell
// @Success 200 {object} model.TagSwagger "成功"
```

接下来你只需要在项目根目录下再次执行 swag init，并在生成成功后再重新启动服务端，就可以查看到最新的效果了，如下：

![image](https://golang2.eddycjy.com/images/ch2/api-desc.jpg)

## 2.4.10 小结

在本章节中，我们简单介绍了 Swagger 和 Swagger 的相关生态圈组件，对所编写的 API 原型新增了响应的 Swagger 注解，在接下来中安装了针对 Go 语言的 Swagger 工具，用于后续的 Swagger 文档生成和使用。

# 2.5 为接口做参数校验

接下来我们将正式进行编码，在进行对应的业务模块开发时，第一步要考虑到的问题的就是如何进行入参校验，我们需要将整个项目，甚至整个团队的组件给定下来，形成一个通用规范，在今天本章节将核心介绍这一块，并完成标签模块的接口的入参校验。

## 2.5.1 validator 介绍

在本项目中我们将使用开源项目 [**go-playground/validator**](https://github.com/go-playground/validator) 作为我们的本项目的基础库，它是一个基于标签来对结构体和字段进行值验证的一个验证器。

那么，我们要单独引入这个库吗，其实不然，因为我们使用的 gin 框架，其内部的模型绑定和验证默认使用的是 [**go-playground/validator**](https://github.com/go-playground/validator) 来进行参数绑定和校验，使用起来非常方便。

在项目根目录下执行命令，进行安装：

```shell
$ go get -u github.com/go-playground/validator/v10
```

## 2.5.2 业务接口校验

接下来我们将正式开始对接口的入参进行校验规则的编写，也就是将校验规则写在对应的结构体的字段标签上，常见的标签含义如下：

| 标签     | 含义                      |
| -------- | ------------------------- |
| required | 必填                      |
| gt       | 大于                      |
| gte      | 大于等于                  |
| lt       | 小于                      |
| lte      | 小于等于                  |
| min      | 最小值                    |
| max      | 最大值                    |
| oneof    | 参数集内的其中之一        |
| len      | 长度要求与 len 给定的一致 |

### 2.5.2.1 标签接口

我们回到项目的 `internal/service` 目录下的 tag.go 文件，针对入参校验增加绑定/验证结构体，在路由方法前写入如下代码：

```go
type CountTagRequest struct {
	Name  string `form:"name" binding:"max=100"`
	State uint8 `form:"state,default=1" binding:"oneof=0 1"`
}

type TagListRequest struct {
	Name  string `form:"name" binding:"max=100"`
	State uint8  `form:"state,default=1" binding:"oneof=0 1"`
}

type CreateTagRequest struct {
	Name      string `form:"name" binding:"required,min=3,max=100"`
	CreatedBy string `form:"created_by" binding:"required,min=3,max=100"`
	State     uint8  `form:"state,default=1" binding:"oneof=0 1"`
}

type UpdateTagRequest struct {
	ID         uint32 `form:"id" binding:"required,gte=1"`
	Name       string `form:"name" binding:"min=3,max=100"`
	State      uint8  `form:"state" binding:"required,oneof=0 1"`
	ModifiedBy string `form:"modified_by" binding:"required,min=3,max=100"`
}

type DeleteTagRequest struct {
	ID uint32 `form:"id" binding:"required,gte=1"`
}
```

在上述代码中，我们主要针对业务接口中定义的的增删改查和统计行为进行了 Request 结构体编写，而在结构体中，应用到了两个 tag 标签，分别是 form 和 binding，它们分别代表着表单的映射字段名和入参校验的规则内容，其主要功能是实现参数绑定和参数检验。

### 2.5.2.2 文章接口

接下来到项目的 `internal/service` 目录下的 article.go 文件，针对入参校验增加绑定/验证结构体。这块与标签模块的验证规则差不多，主要是必填，长度最小、最大的限制，以及要求参数值必须在某个集合内的其中之一，因此不再赘述。

## 2.5.3 国际化处理

### 2.5.3.1 编写中间件

go-playground/validator 默认的错误信息是英文，但我们的错误信息不一定是用的英文，有可能要简体中文，做国际化的又有其它的需求，这可怎么办，在通用需求的情况下，有没有简单又省事的办法解决呢？

如果是简单的国际化需求，我们可以通过中间件配合语言包的方式去实现这个功能，接下来我们在项目的 `internal/middleware` 目录下新建 translations.go 文件，用于编写针对 validator 的语言包翻译的相关功能，新增如下代码：

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant_TW"
	"github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

func Translations() gin.HandlerFunc {
	return func(c *gin.Context) {
		uni := ut.New(en.New(), zh.New(), zh_Hant_TW.New())
		locale := c.GetHeader("locale")
		trans, _ := uni.GetTranslator(locale)
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if ok {
			switch locale {
			case "zh":
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
				break
			case "en":
				_ = en_translations.RegisterDefaultTranslations(v, trans)
				break
			default:
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
				break
			}
			c.Set("trans", trans)
		}

		c.Next()
	}
}
```

在自定义中间件 Translations 中，我们针对 i18n 利用了第三方开源库去实现这块功能，分别如下：

- go-playground/locales：多语言包，是从 CLDR 项目（Unicode 通用语言环境数据存储库）生成的一组多语言环境，主要在 i18n 软件包中使用，该库是与 universal-translator 配套使用的。
- go-playground/universal-translator：通用翻译器，是一个使用 CLDR 数据 + 复数规则的 Go 语言 i18n 转换器。
- go-playground/validator/v10/translations：validator 的翻译器。

而在识别当前请求的语言类别上，我们通过 GetHeader 方法去获取约定的 header 参数 locale，用于判别当前请求的语言类别是 en 又或是 zh，如果有其它语言环境要求，也可以继续引入其它语言类别，因为 go-playground/locales 基本上都支持。

在后续的注册步骤，我们调用 RegisterDefaultTranslations 方法将验证器和对应语言类型的 Translator 注册进来，实现验证器的多语言支持。同时将 Translator 存储到全局上下文中，便于后续翻译时的使用。

### 2.5.3.2 注册中间件

回到项目的 `internal/routers` 目录下的 router.go 文件，新增中间件 Translations 的注册，新增代码如下：

```go
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Translations())
	...
}
```

至此，我们就完成了在项目中的自定义验证器注册、验证器初始化、错误提示多语言的功能支持了。

## 2.5.4 接口校验

我们在项目下的 `pkg/app` 目录新建 form.go 文件，写入如下代码：

```go
import (
	...
	ut "github.com/go-playground/universal-translator"
	val "github.com/go-playground/validator/v10"
)

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

func BindAndValid(c *gin.Context, v interface{}) (bool, ValidErrors) {
	var errs ValidErrors
	err := c.ShouldBind(v)
	if err != nil {
		v := c.Value("trans")
		trans, _ := v.(ut.Translator)
		verrs, ok := err.(val.ValidationErrors)
		if !ok {
			return false, errs
		}

		for key, value := range verrs.Translate(trans) {
			errs = append(errs, &ValidError{
				Key:     key,
				Message: value,
			})
		}

		return false, errs
	}

	return true, nil
}
```

在上述代码中，我们主要是针对入参校验的方法进行了二次封装，在 BindAndValid 方法中，通过 ShouldBind 进行参数绑定和入参校验，当发生错误后，再通过上一步在中间件 Translations 设置的 Translator 来对错误消息体进行具体的翻译行为。

另外我们声明了 ValidError 相关的结构体和类型，对这块不熟悉的读者可能会疑惑为什么要实现其对应的 Error 方法呢，我们简单来看看标准库中 errors 的相关代码，如下：

```go
func New(text string) error {
	return &errorString{text}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

标准库 errors 的 New 方法实现非常简单，errorString 是一个结构体，内含一个 s 字符串，也只有一个 Error 方法，就可以认定为 error 类型，这是为什么呢？这一切的关键都在于 error 接口的定义，如下：

```go
type error interface {
	Error() string
}
```

在 Go 语言中，如果一个类型实现了某个 interface 中的所有方法，那么编译器就会认为该类型实现了此 interface，它们是”一样“的。

## 2.5.5 验证

我们回到项目的 `internal/routers/api/v1` 下的 tag.go 文件，修改获取多个标签的 List 接口，用于验证 validator 是否正常，修改代码如下：

```go
func (t Tag) List(c *gin.Context) {
	param := struct {
		Name  string `form:"name" binding:"max=100"`
		State uint8  `form:"state,default=1" binding:"oneof=0 1"`
	}{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs: %v", errs)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(errs.Errors()...))
		return
	}

	response.ToResponse(gin.H{})
	return
}
```

在命令行中利用 CURL 请求该接口，查看验证结果，如下：

```shell
$ curl -X GET http://127.0.0.1:8000/api/v1/tags\?state\=6
{"code":10000001,"details":["State 必须是[0 1]中的一个"],"msg":"入参错误"}
```

另外你还需要注意到 TagListRequest 的校验规则里其实并没有 required，因此它的校验规则应该是有才校验，没有该入参的话，是默认无校验的，也就是没有 state 参数，也应该可以正常请求，如下：

```shell
$ curl -X GET http://127.0.0.1:8000/api/v1/tags          
{}
```

在 Response 中我们调用的是 `gin.H` 作为返回结果集，因此该输出结果正确。

## 2.5.6 小结

在本章节中，我们介绍了在 gin 框架中如何通过 validator 来进行参数校验，而在一些定制化场景中，我们常常需要自定义验证器，这个时候我们可以通过实现 `binding.Validator` 接口的方式，来替换其自身的 validator：：

```go
// binding/binding.go
type StructValidator interface {
	ValidateStruct(interface{}) error
	Engine() interface{}
}

func setupValidator() error {
	// 将你所自定义的 validator 写入
	binding.Validator = global.Validator
	return nil
}
```

也就是说如果你有定制化需求，也完全可以自己实现一个验证器，效仿我们前面的模式，就可以完全替代 gin 框架原本的 validator 使用了。

而在章节的后半段，我们对业务接口进行了入参校验规则的编写，并且针对错误提示的多语言化问题（也可以理解为一个简单的国际化需求），通过中间件和多语言包的方式进行了实现，在未来如果你有更细致的国际化需求，也可以进一步的拓展。

# 2.6 模块开发：标签管理

在初步完成了业务接口的入参校验的逻辑处理后，接下来我们正式的进入业务模块的业务逻辑开发，在本章节将完成标签模块的接口代码编写，涉及的接口如下：

| 功能         | HTTP 方法 | 路径      |
| ------------ | --------- | --------- |
| 新增标签     | POST      | /tags     |
| 删除指定标签 | DELETE    | /tags/:id |
| 更新指定标签 | PUT       | /tags/:id |
| 获取标签列表 | GET       | /tags     |

## 2.6.1 新建 model 方法

首先我们需要针对标签表进行处理，并在项目的 `internal/model` 目录下新建 tag.go 文件，针对标签模块的模型操作进行封装，并且只与实体产生关系，代码如下：

```go
func (t Tag) Count(db *gorm.DB) (int, error) {
	var count int
	if t.Name != "" {
		db = db.Where("name = ?", t.Name)
	}
	db = db.Where("state = ?", t.State)
	if err := db.Model(&t).Where("is_del = ?", 0).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (t Tag) List(db *gorm.DB, pageOffset, pageSize int) ([]*Tag, error) {
	var tags []*Tag
	var err error
	if pageOffset >= 0 && pageSize > 0 {
		db = db.Offset(pageOffset).Limit(pageSize)
	}
	if t.Name != "" {
		db = db.Where("name = ?", t.Name)
	}
	db = db.Where("state = ?", t.State)
	if err = db.Where("is_del = ?", 0).Find(&tags).Error; err != nil {
		return nil, err
	}
	
	return tags, nil
}

func (t Tag) Create(db *gorm.DB) error {
	return db.Create(&t).Error
}

func (t Tag) Update(db *gorm.DB) error {
	return db.Model(&Tag{}).Where("id = ? AND is_del = ?", t.ID, 0).Update(t).Error
}

func (t Tag) Delete(db *gorm.DB) error {
	return db.Where("id = ? AND is_del = ?", t.Model.ID, 0).Delete(&t).Error
}
```

- Model：指定运行 DB 操作的模型实例，默认解析该结构体的名字为表名，格式为大写驼峰转小写下划线驼峰。若情况特殊，也可以编写该结构体的 TableName 方法用于指定其对应返回的表名。
- Where：设置筛选条件，接受 map，struct 或 string 作为条件。
- Offset：偏移量，用于指定开始返回记录之前要跳过的记录数。
- Limit：限制检索的记录数。
- Find：查找符合筛选条件的记录。
- Updates：更新所选字段。
- Delete：删除数据。
- Count：统计行为，用于统计模型的记录数。

需要注意的是，在上述代码中，我们采取的是将 `db *gorm.DB` 作为函数首参数传入的方式，而在业界中也有另外一种方式，是基于结构体传入的，两者本质上都可以实现目的，读者根据实际情况（使用习惯、项目规范等）进行选用即可，其各有利弊。

## 2.6.2 处理 model 回调

你会发现我们在编写 model 代码时，并没有针对我们的公共字段 created_on、modified_on、deleted_on、is_del 进行处理，难道不是在每一个 DB 操作中进行设置和修改吗？

显然，这在通用场景下并不是最好的方案，因为如果每一个 DB 操作都去设置公共字段的值，那么不仅多了很多重复的代码，在要调整公共字段时工作量也会翻倍。

我们可以采用设置 model callback 的方式去实现公共字段的处理，本项目使用的 ORM 库是 GORM，GORM 本身是提供回调支持的，因此我们可以根据自己的需要自定义 GORM 的回调操作，而在 GORM 中我们可以分别进行如下的回调相关行为：

- 注册一个新的回调。
- 删除现有的回调。
- 替换现有的回调。
- 注册回调的先后顺序。

在本项目中使用到的“替换现有的回调”这一行为，我们打开项目的 `internal/model` 目录下的 model.go 文件，准备开始编写 model 的回调代码，下述所新增的回调代码均写入在 NewDBEngine 方法后。

```go
func NewDBEngine(databaseSetting *setting.DatabaseSettingS) (*gorm.DB, error) {}
func updateTimeStampForCreateCallback(scope *gorm.Scope) {}
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {}
func deleteCallback(scope *gorm.Scope) {}
func addExtraSpaceIfExist(str string) string {}
```

### 2.6.2.1 新增行为的回调

```go
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("CreatedOn"); ok {
			if createTimeField.IsBlank {
				_ = createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok {
			if modifyTimeField.IsBlank {
				_ = modifyTimeField.Set(nowTime)
			}
		}
	}
}
```

- 通过调用 `scope.FieldByName` 方法，获取当前是否包含所需的字段。
- 通过判断 `Field.IsBlank` 的值，可以得知该字段的值是否为空。
- 若为空，则会调用 `Field.Set` 方法给该字段设置值，入参类型为 interface{}，内部也就是通过反射进行一系列操作赋值。

### 2.6.2.2 更新行为的回调

```go
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		_ = scope.SetColumn("ModifiedOn", time.Now().Unix())
	}
}
```

- 通过调用 `scope.Get("gorm:update_column")` 去获取当前设置了标识 `gorm:update_column` 的字段属性。
- 若不存在，也就是没有自定义设置 `update_column`，那么将会在更新回调内设置默认字段 ModifiedOn 的值为当前的时间戳。

### 2.6.2.3 删除行为的回调

```go
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")
		isDelField, hasIsDelField := scope.FieldByName("IsDel")
		if !scope.Search.Unscoped && hasDeletedOnField && hasIsDelField {
			now := time.Now().Unix()
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v,%v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(now),
				scope.Quote(isDelField.DBName),
				scope.AddToVars(1),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
```

- 通过调用 `scope.Get("gorm:delete_option")` 去获取当前设置了标识 `gorm:delete_option` 的字段属性。
- 判断是否存在 `DeletedOn` 和 `IsDel` 字段，若存在则调整为执行 UPDATE 操作进行软删除（修改 DeletedOn 和 IsDel 的值），否则执行 DELETE 进行硬删除。
- 调用 `scope.QuotedTableName` 方法获取当前所引用的表名，并调用一系列方法针对 SQL 语句的组成部分进行处理和转移，最后在完成一些所需参数设置后调用 `scope.CombinedConditionSql` 方法完成 SQL 语句的组装。

### 2.6.2.4 注册回调行为

```go
func NewDBEngine(databaseSetting *setting.DatabaseSettingS) (*gorm.DB, error) {
	...
	db.SingularTable(true)
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.DB().SetMaxIdleConns(databaseSetting.MaxIdleConns)
	db.DB().SetMaxOpenConns(databaseSetting.MaxOpenConns)

	return db, nil
}

func updateTimeStampForCreateCallback(scope *gorm.Scope) {...}
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {...}
func deleteCallback(scope *gorm.Scope) {...}
func addExtraSpaceIfExist(str string) string {...}
```

在最后我们回到 NewDBEngine 方法中，针对上述写的三个 Callback 方法进行回调注册，才能够让我们的应用程序真正的使用上，至此，我们的公共字段处理就完成了。

## 2.6.3 新建 dao 方法

我们在项目的 `internal/dao` 目录下新建 dao.go 文件，写入如下代码：

```go
type Dao struct {
	engine *gorm.DB
}

func New(engine *gorm.DB) *Dao {
	return &Dao{engine: engine}
}
```

接下来在同层级下新建 tag.go 文件，用于处理标签模块的 dao 操作，写入如下代码：

```go
func (d *Dao) CountTag(name string, state uint8) (int, error) {
	tag := model.Tag{Name: name, State: state}
	return tag.Count(d.engine)
}

func (d *Dao) GetTagList(name string, state uint8, page, pageSize int) ([]*model.Tag, error) {
	tag := model.Tag{Name: name, State: state}
	pageOffset := app.GetPageOffset(page, pageSize)
	return tag.List(d.engine, pageOffset, pageSize)
}

func (d *Dao) CreateTag(name string, state uint8, createdBy string) error {
	tag := model.Tag{
		Name:  name,
		State: state,
		Model: &model.Model{CreatedBy: createdBy},
	}

	return tag.Create(d.engine)
}

func (d *Dao) UpdateTag(id uint32, name string, state uint8, modifiedBy string) error {
	tag := model.Tag{
		Name:  name,
		State: state,
		Model: &model.Model{ID: id, ModifiedBy: modifiedBy},
	}

	return tag.Update(d.engine)
}

func (d *Dao) DeleteTag(id uint32) error {
	tag := model.Tag{Model: &model.Model{ID: id}}
	return tag.Delete(d.engine)
}
```

在上述代码中，我们主要是在 dao 层进行了数据访问对象的封装，并针对业务所需的字段进行了处理。

## 2.6.4 新建 service 方法

我们在项目的 `internal/service` 目录下新建 service.go 文件，写入如下代码：

```go
type Service struct {
	ctx context.Context
	dao *dao.Dao
}

func New(ctx context.Context) Service {
	svc := Service{ctx: ctx}
	svc.dao = dao.New(global.DBEngine)
	return svc
}
```

接下来在同层级下新建 tag.go 文件，用于处理标签模块的业务逻辑，写入如下代码：

```go
type CountTagRequest struct {
    Name  string `form:"name" binding:"max=100"`
	State uint8 `form:"state,default=1" binding:"oneof=0 1"`
}

type TagListRequest struct {
	Name  string `form:"name" binding:"max=100"`
	State uint8  `form:"state,default=1" binding:"oneof=0 1"`
}

type CreateTagRequest struct {
	Name      string `form:"name" binding:"required,min=2,max=100"`
	CreatedBy string `form:"created_by" binding:"required,min=2,max=100"`
	State     uint8  `form:"state,default=1" binding:"oneof=0 1"`
}

type UpdateTagRequest struct {
	ID         uint32 `form:"id" binding:"required,gte=1"`
	Name       string `form:"name" binding:"max=100"`
	State      uint8  `form:"state" binding:"oneof=0 1"`
	ModifiedBy string `form:"modified_by" binding:"required,min=2,max=100"`
}

type DeleteTagRequest struct {
	ID uint32 `form:"id" binding:"required,gte=1"`
}

func (svc *Service) CountTag(param *CountTagRequest) (int, error) {
	return svc.dao.CountTag(param.Name, param.State)
}

func (svc *Service) GetTagList(param *TagListRequest, pager *app.Pager) ([]*model.Tag, error) {
	return svc.dao.GetTagList(param.Name, param.State, pager.Page, pager.PageSize)
}

func (svc *Service) CreateTag(param *CreateTagRequest) error {
	return svc.dao.CreateTag(param.Name, param.State, param.CreatedBy)
}

func (svc *Service) UpdateTag(param *UpdateTagRequest) error {
	return svc.dao.UpdateTag(param.ID, param.Name, param.State, param.ModifiedBy)
}

func (svc *Service) DeleteTag(param *DeleteTagRequest) error {
	return svc.dao.DeleteTag(param.ID)
}
```

在上述代码中，我们主要是定义了 Request 结构体作为接口入参的基准，而本项目由于并不会太复杂，所以直接放在了 service 层中便于使用，若后续业务不断增长，程序越来越复杂，service 也冗杂了，可以考虑将抽离一层接口校验层，便于解耦逻辑。

另外我们还在 service 进行了一些简单的逻辑封装，在应用分层中，service 层主要是针对业务逻辑的封装，如果有一些业务聚合和处理可以在该层进行编码，同时也能较好的隔离上下两层的逻辑。

## 2.6.6 新增业务错误码

我们在项目的 `pkg/errcode` 下新建 module_code.go 文件，针对标签模块，写入如下错误代码：

```go
var (
	ErrorGetTagListFail = NewError(20010001, "获取标签列表失败")
	ErrorCreateTagFail  = NewError(20010002, "创建标签失败")
	ErrorUpdateTagFail  = NewError(20010003, "更新标签失败")
	ErrorDeleteTagFail  = NewError(20010004, "删除标签失败")
	ErrorCountTagFail   = NewError(20010005, "统计标签失败")
)
```

## 2.6.7 新增路由方法

我们打开 `internal/routers/api/v1` 项目目录下的 tag.go 文件，写入如下代码：

```go
func (t Tag) List(c *gin.Context) {
	param := service.TagListRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs: %v", errs)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(errs.Errors()...))
		return
	}

	svc := service.New(c.Request.Context())
	pager := app.Pager{Page: app.GetPage(c), PageSize: app.GetPageSize(c)}
	totalRows, err := svc.CountTag(&service.CountTagRequest{Name: param.Name, State: param.State})
	if err != nil {
		global.Logger.Errorf("svc.CountTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorCountTagFail)
		return
	}
	
	tags, err := svc.GetTagList(&param, &pager)
	if err != nil {
		global.Logger.Errorf("svc.GetTagList err: %v", err)
		response.ToErrorResponse(errcode.ErrorGetTagListFail)
		return
	}

	response.ToResponseList(tags, totalRows)
	return
}
```

在上述代码中，我们完成了获取标签列表接口的处理方法，我们在方法中完成了入参校验和绑定、获取标签总数、获取标签列表、 序列化结果集等四大功能板块的逻辑串联和日志、错误处理。

需要注意的是入参校验和绑定的处理代码基本都差不多，因此在后续代码中不再重复，我们继续写入创建标签、更新标签、删除标签的接口处理方法，如下：

```go
func (t Tag) Create(c *gin.Context) {
	param := service.CreateTagRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.CreateTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.CreateTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorCreateTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}

func (t Tag) Update(c *gin.Context) {
	param := service.UpdateTagRequest{ID: convert.StrTo(c.Param("id")).MustUInt32()}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.UpdateTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.UpdateTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorUpdateTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}

func (t Tag) Delete(c *gin.Context) {
	param := service.DeleteTagRequest{ID: convert.StrTo(c.Param("id")).MustUInt32()}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {...}

	svc := service.New(c.Request.Context())
	err := svc.DeleteTag(&param)
	if err != nil {
		global.Logger.Errorf("svc.DeleteTag err: %v", err)
		response.ToErrorResponse(errcode.ErrorDeleteTagFail)
		return
	}

	response.ToResponse(gin.H{})
	return
}
```

## 2.6.8 验证接口

我们重新启动服务，也就是再执行 `go run main.go`，查看启动信息正常后，对标签模块的接口进行验证，请注意，验证示例中的 `{id}`，代指占位符，也就是填写你实际调用中希望处理的标签 ID 即可。

### 2.6.8.1 新增标签

```shell
$ curl -X POST http://127.0.0.1:8000/api/v1/tags -F 'name=Go' -F created_by=eddycjy
{}
$ curl -X POST http://127.0.0.1:8000/api/v1/tags -F 'name=PHP' -F created_by=eddycjy
{}
$ curl -X POST http://127.0.0.1:8000/api/v1/tags -F 'name=Rust' -F created_by=eddycjy
{}
```

### 2.6.8.2 获取标签列表

```shell
$ curl -X GET 'http://127.0.0.1:8000/api/v1/tags?page=1&page_size=2'
{"list":[{"id":1,"created_by":"eddycjy","modified_by":"","created_on":1574493416,"modified_on":1574493416,"deleted_on":0,"is_del":0,"name":"Go 语言","state":1},{"id":2,"created_by":"eddycjy","modified_by":"","created_on":1574493813,"modified_on":1574493813,"deleted_on":0,"is_del":0,"name":"PHP","state":1}],"pager":{"page":1,"page_size":2,"total_rows":3}}

$ curl -X GET 'http://127.0.0.1:8000/api/v1/tags?page=2&page_size=2'
{"list":[{"id":3,"created_by":"eddycjy","modified_by":"","created_on":1574493817,"modified_on":1574493817,"deleted_on":0,"is_del":0,"name":"Rust","state":1}],"pager":{"page":2,"page_size":2,"total_rows":3}}
```

### 2.6.8.3 修改标签

```shell
$ curl -X PUT http://127.0.0.1:8000/api/v1/tags/{id} -F state=0 -F modified_by=eddycjy
{}
```

### 2.6.8.4 删除标签

```shell
$ curl -X DELETE  http://127.0.0.1:8000/api/v1/tags/{id}
{}
```

## 2.6.9 发现问题

在完成了接口的检验后，我们还要确定一下数据库内的数据变更是否正确。在经过一系列的对比后，我们发现在调用修新标签的接口时，通过接口入参，我们是希望将 id 为 1 的标签状态修改为 0，但是在对比后发现数据库内它的状态值居然还是 1，而且 SQL 语句内也没有出现 state 字段的设置，太神奇了，控制台输出的 SQL 语句如下：

```shell
UPDATE `blog_tag` SET `id` = 1, `modified_by` = 'eddycjy', `modified_on` = xxxxx  WHERE `blog_tag`.`id` = 1
```

甚至在我们更进一步其它类似的验证时，发现只要字段是零值的情况下，GORM 就不会对该字段进行变更，这到底是为什么呢？

实际上，这有一个概念上的问题，我们先入为主的认为它一定会变更，其实是不对的，因为在我们程序中使用的是 struct 的方式进行更新操作，而在 GORM 中使用 struct 类型传入进行更新时，GORM 是不会对值为零值的字段进行变更。这又是为什么呢，我们可以猜想，更根本的原因是因为在识别这个结构体中的这个字段值时，很难判定是真的是零值，还是外部传入恰好是该类型的零值，GORM 在这块并没有过多的去做特殊识别。

## 2.6.10 解决问题

修改项目的 `internal/model` 目录下的 tag.go 文件里的 Update 方法，如下：

```go
func (t Tag) Update(db *gorm.DB, values interface{}) error {
	if err := db.Model(t).Where("id = ? AND is_del = ?", t.ID, 0).Updates(values).Error; err != nil {
		return err
	}

	return nil
}
```

修改项目的 `internal/dao` 目录下的 tag.go 文件里的 UpdateTag 方法，如下：

```go
func (d *Dao) UpdateTag(id uint32, name string, state uint8, modifiedBy string) error {
	tag := model.Tag{
		Model: &model.Model{ID: id},
	}
	values := map[string]interface{}{
		"state":       state,
		"modified_by": modifiedBy,
	}
	if name != "" {
		values["name"] = name
	}

	return tag.Update(d.engine, values)
}
```

重新运行程序，请求修改标签接口，如下：

```shell
$ curl -X PUT http://127.0.0.1:8000/api/v1/tags/{id} -F state=0 -F modified_by=eddycjy
{}
```

检查数据是否正常修改，在正确的情况下，该 id 为 1 的标签，modified_by 为 eddycjy，modified_on 应修改为当前时间戳，state 为 0。

## 2.6.11 小结

在本章节中，我们针对 “标签管理” 进行了具体的开发，其中涉及到了 model、dao、service、router 的相关方法以及业务错误码的编写和处理。接下来下一步应当是 “文章管理” 的模块开发，我强烈建议读者根据本章的经验，自行构思设计思路，然后亲自思考和实践，这样子对你未来对实际项目进行开发会有明显帮助。而在开发时，或开发后，如果遇到困难可以参考本书的辅导资料，有包含 “文章管理” 的详细模块开发内容说明。

# 2.7 上传图片和文件服务

在处理文章模块时，你会发现 blog_article 表中的封面图片地址（cover_image_url）是我们直接手动传入的一个虚构地址，那么在实际的应用中，一般不同的架构分层有多种处理方式，例如：由浏览器端调用前端应用，前端应用（客户端）再调用服务端进行上传，第二种是浏览器端直接调用服务端接口上传文件，再调用服务器端的其它业务接口完成业务属性填写。

那么在本章节我们将继续完善功能，实现文章的封面图片上传并用文件服务对外提供静态文件的访问服务，这样子在上传图片后，就可以通过约定的地址访问到该图片资源。

## 2.7.1 新增配置

首先我们打开项目下的 `configs/config.yaml` 配置文件，新增上传相关的配置，如下：

```shell
App:
  ...
  UploadSavePath: storage/uploads
  UploadServerUrl: http://127.0.0.1:8000/static
  UploadImageMaxSize: 5  # MB
  UploadImageAllowExts:
    - .jpg
    - .jpeg
    - .png
```

我们一共新增了四项上传文件所必须的配置项，分别代表的作用如下：

- UploadSavePath：上传文件的最终保存目录。
- UploadServerUrl：上传文件后的用于展示的文件服务地址。
- UploadImageMaxSize：上传文件所允许的最大空间大小（MB）。
- UploadImageAllowExts：上传文件所允许的文件后缀。

接下来我们要在对应的配置结构体上新增上传相关属性，打开项目下的 `pkg/setting/section.go` 新增代码如下：

```go
type AppSettingS struct {
	...
	UploadSavePath       string
	UploadServerUrl      string
	UploadImageMaxSize   int
	UploadImageAllowExts []string
}
```

## 2.7.2 上传文件

接下来我们要编写一个上传文件的工具库，它的主要功能是针对上传文件时的一些相关处理。我们在项目的 `pkg` 目录下新建 `util` 目录，并创建 md5.go 文件，写入如下代码：

```go
func EncodeMD5(value string) string {
	m := md5.New()
	m.Write([]byte(value))

	return hex.EncodeToString(m.Sum(nil))
}
```

该方法用于针对上传后的文件名格式化，简单来讲，将文件名 MD5 后再进行写入，防止直接把原始名称就暴露出去了。接下来我们在项目的 `pkg/upload` 目录下新建 file.go 文件，代码如下：

```go
type FileType int

const TypeImage FileType = iota + 1

func GetFileName(name string) string {
	ext := GetFileExt(name)
	fileName := strings.TrimSuffix(name, ext)
	fileName = util.EncodeMD5(fileName)

	return fileName + ext
}

func GetFileExt(name string) string {
	return path.Ext(name)
}

func GetSavePath() string {
	return global.AppSetting.UploadSavePath
}
```

在上述代码中，我们用到了两个比较常见的语法，首先是我们定义了 FileType 为 int 的类型别名，并且利用 FileType 作为类别标识的基础类型，并 iota 作为了它的初始值，那么 iota 又是什么呢？

实际上，在 Go 语言中 iota 相当于是一个 const 的常量计数器，你也可以理解为枚举值，第一个声明的 iota 的值为 0，在新的一行被使用时，它的值都会自动递增。

当然了，你也可以像代码中那样，在初始的第一个声明时进行手动加一，那么它将会从 1 开始递增。那么为什么我们要在 FileType 类型中使用 iota 的枚举呢，其实本质上是为了后续有其它的需求，能标准化的进行处理，例如：

```go
const (
    TypeImage FileType = iota + 1
    TypeExcel
    TypeTxt
)
```

如果未来再有其它的上传文件类型支持，这么看，是不是就很清晰了呢，不再像以前，你还要手工定义 1, 2, 3, 4….非常麻烦。

另外我们还一共声明了四个文件相关的方法，其作用分别如下：

- GetFileName：获取文件名称，先是通过获取文件后缀并筛出原始文件名进行 MD5 加密，最后返回经过加密处理后的文件名。
- GetFileExt：获取文件后缀，主要是通过调用 `path.Ext` 方法进行循环查找”.“符号，最后通过切片索引返回对应的文化后缀名称。
- GetSavePath：获取文件保存地址，这里直接返回配置中的文件保存目录即可，也便于后续的调整。

在完成了文件相关参数获取的方法后，接下来我们需要编写检查文件的相关方法，因为需要确保在文件写入时它已经达到了必备条件，否则要给出对应的标准错误提示，继续在文件内新增如下代码：

```go
func CheckSavePath(dst string) bool {
	_, err := os.Stat(dst)
	return os.IsNotExist(err)
}

func CheckContainExt(t FileType, name string) bool {
	ext := GetFileExt(name)
	ext = strings.ToUpper(ext)
	switch t {
	case TypeImage:
		for _, allowExt := range global.AppSetting.UploadImageAllowExts {
			if strings.ToUpper(allowExt) == ext {
				return true
			}
		}

	}

	return false
}

func CheckMaxSize(t FileType, f multipart.File) bool {
	content, _ := ioutil.ReadAll(f)
	size := len(content)
	switch t {
	case TypeImage:
		if size >= global.AppSetting.UploadImageMaxSize*1024*1024 {
			return true
		}
	}

	return false
}

func CheckPermission(dst string) bool {
	_, err := os.Stat(dst)
	return os.IsPermission(err)
}
```

- CheckSavePath：检查保存目录是否存在，通过调用 `os.Stat` 方法获取文件的描述信息 FileInfo，并调用 `os.IsNotExist` 方法进行判断，其原理是利用 `os.Stat` 方法所返回的 error 值与系统中所定义的 `oserror.ErrNotExist` 进行判断，以此达到校验效果。
- CheckPermission：检查文件权限是否足够，与 `CheckSavePath` 方法原理一致，是利用 `oserror.ErrPermission` 进行判断。
- CheckContainExt：检查文件后缀是否包含在约定的后缀配置项中，需要的是所上传的文件的后缀有可能是大写、小写、大小写等，因此我们需要调用 `strings.ToUpper` 方法统一转为大写（固定的格式）来进行匹配。
- CheckMaxSize：检查文件大小是否超出最大大小限制。

在完成检查文件的一些必要操作后，我们就可以涉及文件写入/创建的相关操作，继续在文件内新增如下代码：

```go
func CreateSavePath(dst string, perm os.FileMode) error {
	err := os.MkdirAll(dst, perm)
	if err != nil {
		return err
	}

	return nil
}

func SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
```

- CreateSavePath：创建在上传文件时所使用的保存目录，在方法内部调用的 `os.MkdirAll` 方法，该方法将会以传入的 `os.FileMode` 权限位去递归创建所需的所有目录结构，若涉及的目录均已存在，则不会进行任何操作，直接返回 nil。
- SaveFile：保存所上传的文件，该方法主要是通过调用 `os.Create` 方法创建目标地址的文件，再通过 `file.Open` 方法打开源地址的文件，结合 `io.Copy` 方法实现两者之间的文件内容拷贝。

## 2.7.3 新建 service 方法

我们将上一步所编写的上传文件工具库与我们具体的业务接口结合起来，我们在项目下的 `internal/service` 目录新建 upload.go 文件，写入如下代码：

```go
type FileInfo struct {
	Name      string
	AccessUrl string
}

func (svc *Service) UploadFile(fileType upload.FileType, file multipart.File, fileHeader *multipart.FileHeader) (*FileInfo, error) {
	fileName := upload.GetFileName(fileHeader.Filename)
	if !upload.CheckContainExt(fileType, fileName) {
		return nil, errors.New("file suffix is not supported.")
	}
	if upload.CheckMaxSize(fileType, file) {
		return nil, errors.New("exceeded maximum file limit.")
	}

	uploadSavePath := upload.GetSavePath()
	if upload.CheckSavePath(uploadSavePath) {
		if err := upload.CreateSavePath(uploadSavePath, os.ModePerm); err != nil {
			return nil, errors.New("failed to create save directory.")
		}
	}
	if upload.CheckPermission(uploadSavePath) {
		return nil, errors.New("insufficient file permissions.")
	}

	dst := uploadSavePath + "/" + fileName
	if err := upload.SaveFile(fileHeader, dst); err != nil {
		return nil, err
	}

	accessUrl := global.AppSetting.UploadServerUrl + "/" + fileName
	return &FileInfo{Name: fileName, AccessUrl: accessUrl}, nil
}
```

我们在 UploadFile Service 方法中，主要是通过获取文件所需的基本信息，接着对其进行业务所需的文件检查（文件大小是否符合需求、文件后缀是否达到要求），并且判断在写入文件前对否具备必要的写入条件（目录是否存在、权限是否足够），最后在检查通过后再进行真正的写入文件操作。

## 2.7.4 新增业务错误码

在项目的 `pkg/errcode` 下的 module_code.go 文件，针对上传模块，新增如下错误代码：

```go
var (
	...
	ErrorUploadFileFail = NewError(20030001, "上传文件失败")
)
```

## 2.7.5 新增路由方法

接下来需要编写上传文件的路由方法，将整套上传逻辑给串联起来，我们在项目的 `internal/routers` 目录下新建 upload.go 文件，代码如下：

```go
type Upload struct{}

func NewUpload() Upload {
	return Upload{}
}


func (u Upload) UploadFile(c *gin.Context) {
	response := app.NewResponse(c)
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	fileType := convert.StrTo(c.PostForm("type")).MustInt()
	if fileHeader == nil || fileType <= 0 {
		response.ToErrorResponse(errcode.InvalidParams)
		return
	}

	svc := service.New(c.Request.Context())
	fileInfo, err := svc.UploadFile(upload.FileType(fileType), file, fileHeader)
	if err != nil {
		global.Logger.Errorf(c, "svc.UploadFile err: %v", err)
		response.ToErrorResponse(errcode.ErrorUploadFileFail.WithDetails(err.Error()))
		return
	}

	response.ToResponse(gin.H{
		"file_access_url": fileInfo.AccessUrl,
	})
}
```

在上述代码中，我们通过 `c.Request.FormFile` 读取入参 file 字段的上传文件信息，并利用入参 type 字段作为所上传文件类型的确立依据（也可以通过解析上传文件后缀来确定文件类型），最后通过入参检查后进行 Serivce 的调用，完成上传和文件保存，返回文件的展示地址。

至此，业务接口的编写就完成了，下一步我们需要添加路由，让外部能够访问到该接口，依旧是在 `internal/routers` 目录下的 router.go 文件，我们在之中新增上传文件的对应路由，如下：

```go
func NewRouter() *gin.Engine {
	...
	upload := api.NewUpload()
	r.POST("/upload/file", upload.UploadFile)
	apiv1 := r.Group("/api/v1"){...}
}
```

我们新增了 POST 方法的 `/upload/file` 路由，并调用其 upload.UploadFile 方法来提供接口的方法响应，至此整体的路由到业务接口的联通就完成了。

## 2.7.6 验证接口

```shell
$ curl -X POST http://127.0.0.1:8000/upload/file -F file=@{file_path} -F type=1
{
    "file_access_url": "http://127.0.0.1:8000/static/379efdddb61250a2c589e4c28744c4d9.jpeg"
}
```

检查接口返回是否与期望的一致，主体是由 UploadServerUrl 与加密后的文件名称相结合。

## 2.7.7 文件服务

在进行接口的返回结果校验时，你会发现上小节中 file_access_url 这个地址压根就无法访问到对应的文件资源，检查文件资源也确实存在 `storage/uploads` 目录下，这是怎么回事呢？

实际上是需要设置文件服务去提供静态资源的访问，才能实现让外部请求本项目 HTTP Server 时同时提供静态资源的访问，实际上在 gin 中实现 File Server 是非常简单的，我们需要在 NewRouter 方法中，新增如下路由：

```go
func NewRouter() *gin.Engine {
	...
	r.POST("/upload/file", upload.UploadFile)
	r.StaticFS("/static", http.Dir(global.AppSetting.UploadSavePath))
	apiv1 := r.Group("/api/v1"){...}
	return r
}
```

新增 StaticFS 路由完毕后，重新重启应用程序，再次访问 file_access_url 所输出的地址就可以查看到刚刚上传的静态文件了。

## 2.7.8 发生了什么

看到这里你可能会疑惑，为什么设置一个 r.StaticFS 的路由，就可以拥有一个文件服务，并且能够提供静态资源的访问呢，真是神奇。我们可以反过来思考，既然能够读取到文件的展示，那么就是在访问 `$HOST/static` 时，应用程序会读取到 `blog-service/storage/uploads` 下的文件。我们可以看看 StaticFS 方法到底做了什么事，方法原型如下：

```go
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.returnObj()
}
```

首先可以看到在暴露的 URL 中程序禁止了“*”和“:”符号的使用，然后通过 `createStaticHandler` 创建了静态文件服务，其实质最终调用的还是 `fileServer.ServeHTTP` 和对应的处理逻辑，如下：

```go
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	_, nolisting := fs.(*onlyfilesFS)
	return func(c *Context) {
		if nolisting {
			c.Writer.WriteHeader(404)
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
```

在 createStaticHandler 方法中，我们可以留意下 `http.StripPrefix` 方法的调用，实际上在静态文件服务中很常见，它主要作用是从请求 URL 的路径中删除给定的前缀，然后返回一个 Handler。

另外我们在 StaticFS 方法中看到 `urlPattern := path.Join(relativePath, "/*filepath")` 的代码块，而 `/*filepath` 就非常迷惑了，它是什么，又有什么作用呢。我们通过语义可得知它是路由的处理逻辑，而 gin 的路由是基于 httprouter 的，通过查阅文档可以得到如下信息：

```shell
Pattern: /src/*filepath

 /src/                     match
 /src/somefile.go          match
 /src/subdir/somefile.go   match
```

简单来讲，`*filepath` 将会匹配所有文件路径，但是前提是 `*filepath` 标识符必须在 Pattern 的最后。

## 2.7.9 小结

在本章节中我们针对文章所需的封面图，实现了上传图片接口和静态资源文件服务的功能，从中你可以学习到常见的文件处理操作以及文件服务访问的实现方式。另外在实际项目中，你有一点需要注意，你应当将应用服务和文件服务给拆分开来，因为从安全角度来讲，文件资源不应当与应用资源摆放在一起，否则会有风险，又或是直接采用市面上的 OSS 也是可以的。

# 2.8 对接口进行访问控制

在完成了相关的业务接口的开发后，你正打算放到服务器上给其他同事查看时，你又想到了一个问题，这些 API 接口，没有鉴权功能，那就是所有知道地址的人都可以请求该项目的 API 接口和 Swagger 文档，甚至有可能会被网络上的端口扫描器扫描到后滥用，这非常的不安全，怎么办呢。实际上，我们应该要考虑做纵深防御，对 API 接口进行访问控制。

目前市场上比较常见的两种 API 访问控制方案，分别是 OAuth 2.0 和 JWT，但实际上这两者并不能直接的进行对比，因为它们是两个完全不同的东西，对应的应用场景也不一样，我们可以先大致了解，如下：

- OAuth 2.0：本质上是一个授权的行业标准协议，提供了一整套的授权机制的指导标准，常用于使用第三方登陆的情况，像是你在网站登录时，会有提供其它第三方站点（例如用微信、QQ、Github 账号）关联登陆的，往往就是用 OAuth 2.0 的标准去实现的。并且 OAuth 2.0 会相对重一些，常常还会授予第三方应用去获取到我们对应账号的个人基本信息等等。
- JWT：与 OAuth 2.0 完全不同，它常用于前后端分离的情况，能够非常便捷的给 API 接口提供安全鉴权，因此在本章节我们采用的就是 JWT 的方式，来实现我们的 API 访问控制功能。

## 2.8.1 JWT 是什么

JSON Web 令牌（JWT）是一个开放标准（RFC7519），它定义了一种紧凑且自包含的方式，用于在各方之间作为 JSON 对象安全地传输信息。 由于此信息是经过数字签名的，因此可以被验证和信任。 可以使用使用 RSA 或 ECDSA 的公用/专用密钥对对 JWT 进行签名，其格式如下：

![image](https://golang2.eddycjy.com/images/ch2/jwt-format.jpg)

JSON Web 令牌（JWT）是由紧凑的形式三部分组成，这些部分由点 “.“ 分隔，组成为 ”xxxxx.yyyyy.zzzzz“ 的格式，三个部分分别代表的意义如下：

- Header：头部。
- Payload：有效载荷。
- Signature：签名。

### 2.8.1.1 Header

Header（头部）通常由两部分组成，分别是令牌的类型和所使用的签名算法（HMAC SHA256、RSA 等），其会组成一个 JSON 对象用于描述其元数据，例如：

```shell
{
  "alg": "HS256",
  "typ": "JWT"
}
```

在上述 JSON 中 alg 字段表示所使用的签名算法，默认是 HMAC SHA256（HS256），而 type 字段表示所使用的令牌类型，我们使用的 JWT 令牌类型，在最后会对上面的 JSON 对象进行 base64UrlEncode 算法进行转换成为 JWT 的第一部分。

### 2.8.1.2 Payload

Payload（有效负载）也是一个 JSON 对象，主要存储在 JWT 中实际传输的数据，如下：

```shell
{
  "sub": "1234567890",
  "name": "John Doe",
  "admin": true
}
```

- aud（Audience）：受众，也就是接受 JWT 的一方。
- exp（ExpiresAt）：所签发的 JWT 过期时间，过期时间必须大于签发时间。
- jti（JWT Id）：JWT 的唯一标识。
- iat（IssuedAt）：签发时间
- iss（Issuer）：JWT 的签发者。
- nbf（Not Before）：JWT 的生效时间，如果未到这个时间则为不可用。
- sub（Subject）：主题

同样也会对该 JSON 对象进行 base64UrlEncode 算法将其转换为 JWT Token 的第二部分。

这时候你需要注意一个问题点，也就是 JWT 在转换时用的 base64UrlEncode 算法，也就是它是可逆的，因此一些敏感信息请不要放到 JWT 中，若有特殊情况一定要放，也应当进行一定的加密处理。

### 2.8.1.3 Signature

Signature（签名）部分是对前面两个部分组合（Header+Payload）进行约定算法和规则的签名，而签名将会用于校验消息在整个过程中有没有被篡改，并且对有使用私钥进行签名的令牌，它还可以验证 JWT 的发送者是否它的真实身份。

在签名的生成上，在应用程序指定了密钥（secret）后，会使用传入的指定签名算法（默认是 HMAC SHA256），然后通过下述的签名方式来完成 Signature（签名）部分的生成，如下：

```shell
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret)
```

因此我们可以看出 JWT 的第三部分是由 Header、Payload 以及 Secret 的算法组成而成的，因此它最终可达到用于校验消息是否被篡改的作用之一，因为如果一旦被篡改，Signature 就会无法对上。

### 2.8.1.4 Base64UrlEncode

我们可以在上述章节中不断的看到 Header、Payload 以及 Signature 的签名算法均使用到了 Base64UrlEncode 函数，它究竟是什么呢。

实际上 Base64UrlEncode 是 Base64 算法的变种，为什么要变呢，原因是在业界中我们经常可以看到 JWT 令牌会被放入 Header 或 Query Param 中（也就是 URL）。

而在 URL 中，一些个别字符是有特殊意义的，例如：“+”、“/”、“=” 等等，因此在 Base64UrlEncode 算法中，会对其进行替换，例如：“+” 替换为 “-”、“/” 替换成 “_”、“=” 会被进行忽略处理，以此来保证 JWT 令牌的在 URL 中的可用性和准确性。

## 2.8.2 JWT 的使用场景

通常会先在内部约定好 JWT 令牌的交流方式，像是存储在 Header、Query Param、Cookie、Session 都有，但最常见的是存储在 Header 中。然后服务端提供一个获取 JWT 令牌的接口方法，返回而客户端去使用，在客户端请求其余的接口时需要带上所签发的 JWT 令牌，然后服务端接口也会到约定位置上获取 JWT 令牌来进行鉴权处理，以此流程来鉴定是否合法。

## 2.8.3 安装 JWT

接下来开始对项目进行 JWT 的相关处理，首先我们需要拉取 jwt-go 库，该库提供了 JSON Web 令牌（JWT）的 Go 实现，能够便捷的提供 JWT 支持，不需要我们自己亲自去实现，执行如下命令：

```shell
$ go get -u github.com/dgrijalva/jwt-go@v3.2.0
```

## 2.8.4 配置 JWT

### 2.8.4.1 创建认证表

在介绍 JWT 和其使用场景时，我们知道了实际上需要一个服务端的接口来提供 JWT 令牌的签发，并且可以将自定义的私有信息存入其中，那么我们必然需要一个地方来存储签发的凭证，否则谁来都签发，似乎不大符合实际的业务需求，因此我们要创建一个新的数据表，用于存储签发的认证信息，表 SQL 语句如下：

```shell
CREATE TABLE `blog_auth` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `app_key` varchar(20) DEFAULT '' COMMENT 'Key',
  `app_secret` varchar(50) DEFAULT '' COMMENT 'Secret',
  # 此处请写入公共字段
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证管理';
```

上述表 SQL 语句的主要作用是创建了一张名为 blog_auth 的表，其核心是 app_key 和 app_secret 字段，用于签发的认证信息，接下来我们默认插入一条认证的 SQL 语句（你也可以做一个接口），便于我们认证接口的后续使用，插入的 SQL 语句如下：

```shell
INSERT INTO `blog_service`.`blog_auth`(`id`, `app_key`, `app_secret`, `created_on`, `created_by`, `modified_on`, `modified_by`, `deleted_on`, `is_del`) VALUES (1, 'eddycjy', 'go-programming-tour-book', 0, 'eddycjy', 0, '', 0, 0);
```

该条语句的主要作用是新增了一条 app_key 为 eddycjy 以及 app_secret 为 go-programming-tour-book 的数据。

### 2.8.4.2 新建 model 对象

接下来打开项目的 `internal/model` 目录下的 auth.go 文件，写入对应刚刚新增的 blog_auth 表的数据模型，如下：

```go
type Auth struct {
	*Model
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

func (a Auth) TableName() string {
	return "blog_auth"
}
```

### 2.8.4.2 初始化配置

接下来需要针对 JWT 的一些相关配置进行设置，我们打开项目的 `configs/config.yaml` 配置文件，写入新的配置项，如下：

```shell
JWT:
  Secret: eddycjy
  Issuer: blog-service
  Expire: 7200
```

然后对 JWT 的配置进行初始化操作，我们打开项目的启动文件 main.go，修改其 setupSetting 方法，如下：

```go
func setupSetting() error {
   ...
   err = s.ReadSection("JWT", &global.JWTSetting)
   if err != nil {
      return err
   }

   global.JWTSetting.Expire *= time.Second
   ...
}
```

在上述配置中，我们设置了 JWT 令牌的 Secret（密钥）为 eddycjy，签发者（Issuer）是 blog-service，有效时间（Expire）为 7200 秒，这里需要注意的是 Secret 千万不要暴露给外部，只能有服务端知道，否则是可以解密出来的，非常危险。

## 2.8.5 处理 JWT 令牌

虽然 jwt-go 库能够帮助我们快捷的处理 JWT 令牌相关的行为，但是我们还是需要根据我们的项目特性对其进行设计的，简单来讲，就是组合其提供的 API，设计我们的鉴权场景。

接下来我们打开项目目录 `pkg/app` 并创建 jwt.go 文件，写入第一部分的代码：

```go
type Claims struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
	jwt.StandardClaims
}

func GetJWTSecret() []byte {
	return []byte(global.JWTSetting.Secret)
}
```

这块主要涉及 JWT 的一些基本属性，第一个是 GetJWTSecret 方法，用于获取该项目的 JWT Secret，目前我们是直接使用配置所配置的 Secret，第二个是 Claims 结构体，分为两大块，第一块是我们所嵌入的 AppKey 和 AppSecret，用于我们自定义的认证信息，第二块是 `jwt.StandardClaims` 结构体，它是 jwt-go 库中预定义的，也是 JWT 的规范，其涉及字段如下：

```go
type StandardClaims struct {
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	Id        string `json:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty"`
}
```

我想你一看就明白了，它对应的其实是 2.8.1.2 章节中 Payload 的相关字段，这些字段都是非强制性但官方建议使用的预定义权利要求，能够提供一组有用的，可互操作的约定。

接下来我们继续在 jwt.go 文件中写入第二部分代码，如下：

```go
func GenerateToken(appKey, appSecret string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(global.JWTSetting.Expire)
	claims := Claims{
		AppKey:    util.EncodeMD5(appKey),
		AppSecret: util.EncodeMD5(appSecret),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    global.JWTSetting.Issuer,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(GetJWTSecret())
	return token, err
}
```

在 GenerateToken 方法中，它承担了整个流程中比较重要的职责，也就是生成 JWT Token 的行为，主体的函数流程逻辑是根据客户端传入的 AppKey 和 AppSecret 以及在项目配置中所设置的签发者（Issuer）和过期时间（ExpiresAt），根据指定的算法生成签名后的 Token。这其中涉及两个的内部方法，如下：

- jwt.NewWithClaims：根据 Claims 结构体创建 Token 实例，它一共包含两个形参，第一个参数是 SigningMethod，其包含 SigningMethodHS256、SigningMethodHS384、SigningMethodHS512 三种 crypto.Hash 加密算法的方案。第二个参数是 Claims，主要是用于传递用户所预定义的一些权利要求，便于后续的加密、校验等行为。
- tokenClaims.SignedString：生成签名字符串，根据所传入 Secret 不同，进行签名并返回标准的 Token。

接下来我们继续在 jwt.go 文件中写入第三部分代码，如下：

```go
func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return GetJWTSecret(), nil
	})
    if err != nil {
	    return nil, err
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
```

在 ParseToken 方法中，它主要的功能是解析和校验 Token，承担着与 GenerateToken 相对的功能，其函数流程主要是解析传入的 Token，然后根据 Claims 的相关属性要求进行校验。这其中涉及两个的内部方法，如下：

- ParseWithClaims：用于解析鉴权的声明，方法内部主要是具体的解码和校验的过程，最终返回 `*Token`。
- Valid：验证基于时间的声明，例如：过期时间（ExpiresAt）、签发者（Issuer）、生效时间（Not Before），需要注意的是，如果没有任何声明在令牌中，仍然会被认为是有效的。

至此我们就完成了 JWT 令牌的生成、解析、校验的方法编写，我们会在后续的应用中间件中对其进行调用，使其能够在应用程序中将一整套的动作给串联起来。

## 2.8.6 获取 JWT 令牌

### 2.8.6.1 新建 model 方法

在前面的章节中，我们为了记录 JWT 令牌的认证信息，新增了 blog_auth 表，因此我们需要新增同样的 model 行为，打开项目目录 `internal/model` 的 auth.go 文件，写入如下代码：

```go
func (a Auth) Get(db *gorm.DB) (Auth, error) {
	var auth Auth
	db = db.Where("app_key = ? AND app_secret = ? AND is_del = ?", a.AppKey, a.AppSecret, 0)
	err := db.First(&auth).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return auth, err
	}

	return auth, nil
}
```

上述方法主要是用于服务端在获取客户端所传入的 app_key 和 app_secret 后，根据所传入的认证信息进行获取，以此判别是否真的存在这一条数据。

### 2.8.6.2 新建 dao 方法

接下来打开项目目录 `internal/dao` 的 auth.go 文件，新增针对获取认证信息的方法，写入如下代码：

```go
func (d *Dao) GetAuth(appKey, appSecret string) (model.Auth, error) {
	auth := model.Auth{AppKey: appKey, AppSecret: appSecret}
	return auth.Get(d.engine)
}
```

### 2.8.6.3 新建 service 方法

接下来打开 `internal/service` 的 auth.go 文件，针对一些相应的基本逻辑进行处理，写入如下代码：

```go
type AuthRequest struct {
	AppKey    string `form:"app_key" binding:"required"`
	AppSecret string `form:"app_secret" binding:"required"`
}

func (svc *Service) CheckAuth(param *AuthRequest) error {
	auth, err := svc.dao.GetAuth(param.AppKey, param.AppSecret)
	if err != nil {
		return err
	}

	if auth.ID > 0 {
		return nil
	}

	return errors.New("auth info does not exist.")
}
```

在上述代码中，我们声明了 AuthRequest 结构体用于接口入参的校验，AppKey 和 AppSecret 都设置为了必填项，在 CheckAuth 方法中，我们使用客户端所传入的认证信息作为筛选条件获取数据行，以此根据是否取到认证信息 ID 来进行是否存在的判定。

### 2.8.6.4 新增路由方法

接下来打开项目目录 `internal/routers/api` 的 auth.go 文件，写入如下代码：

```go
func GetAuth(c *gin.Context) {
	param := service.AuthRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, &param)
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs: %v", errs)
		response.ToErrorResponse(errcode.InvalidParams.WithDetails(errs.Errors()...))
		return
	}

	svc := service.New(c.Request.Context())
	err := svc.CheckAuth(&param)
	if err != nil {
		global.Logger.Errorf("svc.CheckAuth err: %v", err)
		response.ToErrorResponse(errcode.UnauthorizedAuthNotExist)
		return
	}

	token, err := app.GenerateToken(param.AppKey, param.AppSecret)
	if err != nil {
		global.Logger.Errorf("app.GenerateToken err: %v", err)
		response.ToErrorResponse(errcode.UnauthorizedTokenGenerate)
		return
	}

	response.ToResponse(gin.H{
		"token": token,
	})
}
```

这块的逻辑主要是校验及获取入参后，绑定并获取到的 app_key 和 app_secrect 进行数据库查询，检查认证信息是否存在，若存在则进行 Token 的生成并返回。

接下来我们打开项目目录 `internal/routers` 的 router.go 文件，新增 `auth` 相关路由，如下：

```go
func NewRouter() *gin.Engine {
	...
	r.POST("/auth", api.GetAuth)
	...
}
```

至此，就完成了获取 Token 的整套流程。

### 2.8.6.7 接口验证

在完成后，我们需要重新启动服务，并且验证获取 Token 的接口是否正常，如下：

```shell
$ curl -X POST \
  'http://127.0.0.1:8000/auth' \
  -H 'app_key: eddycjy' \
  -H 'app_secret: go-programming-tour-book'
  
{"token":"eyJhbGciOiJIUxxx.eyJhcHBfa2V5Ixxx.omW-x23ZVG5I7cjoWTLVUYxxx..."}
```

## 2.8.7 处理应用中间件

### 2.8.7.1 编写 JWT 中间件

在完成了获取 Token 的接口后，你可能会疑惑，能获取了 Token 了，但是对于其它的业务接口，它还没产生任何作用，那我们应该如何将整个应用流程给串起来呢。那么涉及特定类别的接口统一处理，那必然是选择应用中间件的方式，接下来我们打开项目目录 `internal/middleware` 并新建 jwt.go 文件，写入如下代码：

```go
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			token string
			ecode = errcode.Success
		)
		if s, exist := c.GetQuery("token"); exist {
			token = s
		} else {
			token = c.GetHeader("token")
		}
		if token == "" {
			ecode = errcode.InvalidParams
		} else {
			_, err := app.ParseToken(token)
			if err != nil {
				switch err.(*jwt.ValidationError).Errors {
				case jwt.ValidationErrorExpired:
					ecode = errcode.UnauthorizedTokenTimeout
				default:
					ecode = errcode.UnauthorizedTokenError
				}
			}
		}

		if ecode != errcode.Success {
			response := app.NewResponse(c)
			response.ToErrorResponse(ecode)
			c.Abort()
			return
		}

		c.Next()
	}
}
```

在上述代码中，我们通过 GetHeader 方法从 Header 中获取 token 参数，并调用 ParseToken 对其进行解析，再根据返回的错误类型进行断言判定。

### 2.8.7.2 接入 JWT 中间件

在完成了 JWT 的中间件编写后，我们需要将其接入到应用流程中，但是需要注意的是，并非所有的接口都需要用到 JWT 中间件，因此我们需要利用 gin 中的分组路由的概念，只针对 apiv1 的路由分组进行 JWT 中间件的引用，也就是只有 apiv1 路由分组里的路由方法会受此中间件的约束，如下：

```go
func NewRouter() *gin.Engine {
	...
	apiv1 := r.Group("/api/v1")
	apiv1.Use(middleware.JWT()){...}
	return r
}
```

## 2.8.7.3 验证接口

### 2.8.7.4 没有获取 Token

```shell
$ curl -X GET http://127.0.0.1:8000/api/v1/tags
{"code":10000001,"msg":"入参错误"}
```

### 2.8.7.3.2 Token 错误

```shell
$ curl -X GET http://127.0.0.1:8000/api/v1/tags -H 'token: eyJhbGciOiJIUzI1NiIsInRxxx'
{"code":10000004,"msg":"鉴权失败，Token 错误"}
```

### 2.8.7.3.3 Token 超时

```shell
$ curl -X GET http://127.0.0.1:8000/api/v1/tags -H 'token: eyJhbGciOiJIUzI1NiIsInRxxx'
{"code":10000005,"msg":"鉴权失败，Token 超时"}
```

## 2.8.8 思考

我们通过本章节的学习，可以得知 JWT 令牌的内容是非严格加密的，大体上只是进行 base64UrlEncode 的处理，也就是对 JWT 令牌机制有一定了解的人可以进行反向解密，我们在这里可以做一个演示，首先你先调用 `/auth` 接口获取一个全新 token，例如：

```shell
{
    "token": "eyJhbGci...kpXVCJ9.eyJhcHBfa....DM5MTcsImlzcyI6ImJsb2ctc2VydmljZSJ9.phkGM6...Df1Cc8UC0"
}
```

接下来针对你新获取的 Token 值，只需要手动复制中间那一段（也就是 Payload），然后编写一个测试 Demo 来进行 base64 的解码，Demo 代码如下：

```go
func main() {
	payload, _ := base64.StdEncoding.DecodeString("eyJhcHBfa....DM5MTcsImlzcyI6ImJsb2ctc2VydmljZSJ9")
	log.Println(string(payload))
}
```

最终的输出结果，如下：

```shell
{"app_key":"27566...ccf1","app_secret":"7c97...f4","exp":1576403917,"iss":"blog-service"}
```

你可以看到，假设有人拦截到你的 Token 后，是可以通过你的 Token 去解密并获取到你的 Payload 信息，也就是至少你在在 Payload 中不应该明文存储重要的信息，若非要存，就必须要进行不可逆加密，这样子才可以确保一定的安全性。

同时你也可以发现，过期时间（ExpiresAt）是存储在 Payload 中的，也就是 JWT 令牌一旦签发，在没有做特殊逻辑的情况下，过期时间是不可以再度变更的，因此务必根据自己的实际项目情况进行设计和思考。

# 2.9 应用中间件

完成了接口的访问控制后，心中的一块大石终于落地了，你在开发服务器上将这个项目运行了起来，等着另外一位同事和你对接你所编写的后端接口后便愉快的先下班了。

但结果第二天你一来，该同事非常苦恼的和你说，你的接口，怎么调一下就出问题了，你大为震惊，详细的咨询了是几时调用的接口，调用的接口是哪个，入参又是什么？这时候更无奈的问题出现了，该同事只记得好像大概是晚上 9 点多，入参忘记记录了，它的调试工具上也是密密麻麻的访问记录，根本就分不清楚是哪一条入参记录，它只隐隐约约的记得是某一个接口。

这时候的你，想着去开发服务器上看看访问情况，结果发现，你默认使用的是 gin 的 Logging 和 Recovery，也就是在控制台上输出一些访问和异常记录，但很尴尬的是，它并没有成功记录到你所需要的一些数据，这样子你就无法及时的进行复现和响应，更别说现在还没进行多服务间内部调用和压力测试了。

以上问题，在一个项目的雏形初期很常见，实际上针对不同的环境，我们应该进行一些特殊的调整，而往往这些都是有规律可依的，一些常用的应用中间件就可以妥善的解决这些问题，接下来在这篇文章中我们去编写一些在项目中比较常见的应用中间件。

## 2.9.1 访问日志记录

在出问题时，我们常常会需要去查日志，那么除了查错误日志、业务日志以外，还有一个很重要的日志类别，就是访问日志，从功能上来讲，它最基本的会记录每一次请求的请求方法、方法调用开始时间、方法调用结束时间、方法响应结果、方法响应结果状态码，更进一步的话，会记录 RequestId、TraceId、SpanId 等等附加属性，以此来达到日志链路追踪的效果，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-access-log.jpg)

但是在正式开始前，你又会遇到一个问题，你没办法非常直接的获取到方法所返回的响应主体，这时候我们需要巧妙利用 Go interface 的特性，实际上在写入流时，调用的是 http.ResponseWriter，如下：

```go
type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}
```

那么我们只需要写一个针对访问日志的 Writer 结构体，实现我们特定的 Write 方法就可以解决无法直接取到方法响应主体的问题了。我们打开项目目录 `internal/middleware` 并创建 access_log.go 文件，写入如下代码：

```go
type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}
```

我们在 AccessLogWriter 的 Write 方法中，实现了双写，因此我们可以直接通过 AccessLogWriter 的 body 取到值，接下来我们继续编写访问日志的中间件，写入如下代码：

```go
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyWriter := &AccessLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter

		beginTime := time.Now().Unix()
		c.Next()
		endTime := time.Now().Unix()

		fields := logger.Fields{
			"request":  c.Request.PostForm.Encode(),
			"response": bodyWriter.body.String(),
		}
		global.Logger.WithFields(fields).Infof("access log: method: %s, status_code: %d, begin_time: %d, end_time: %d",
			c.Request.Method,
			bodyWriter.Status(),
			beginTime,
			endTime,
		)
	}
}
```

在 AccessLog 方法中，我们初始化了 AccessLogWriter，将其赋予给当前的 Writer 写入流（可理解为替换原有），并且通过指定方法得到我们所需的日志属性，最终写入到我们的日志中去，其中涉及到了如下信息：

- method：当前的调用方法。
- request：当前的请求参数。
- response：当前的请求结果响应主体。
- status_code：当前的响应结果状态码。
- begin_time/end_time：调用方法的开始时间，调用方法结束的结束时间。

## 2.9.2 异常捕获处理

在异常造成的恐慌发生时，你一定不在现场，因为你不能随时随地的盯着控制台，在常规手段下你也不知道它几时有可能发生，因此对于异常的捕获和及时的告警通知是非常重要的，而发现这些可能性的手段有非常多，我们本次采取的是最简单的捕获和告警通知，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-recovery-mail.jpg)

### 2.9.2.1 自定义 Recovery

在前文中我们看到 gin 本身已经自带了一个 Recovery 中间件，但是在项目中，我们需要针对我们的公司内部情况或生态圈定制 Recovery 中间件，确保异常在被正常捕抓之余，要及时的被识别和处理，因此自定义一个 Recovery 中间件是非常有必要的，如下：

```go
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				global.Logger.WithCallersFrames().Errorf("panic recover err: %v", err)
				app.NewResponse(c).ToErrorResponse(errcode.ServerError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

### 2.9.2.2 邮件报警处理

另外我们在实现 Recovery 的同时，需要实现一个简单的邮件报警功能，确保出现 Panic 后，在捕抓之余能够通过邮件报警来及时的通知到对应的负责人。

#### 2.9.2.2.1 安装

首先在项目目录下执行安装命令，如下：

```shell
go get -u gopkg.in/gomail.v2
```

Gomail 是一个用于发送电子邮件的简单又高效的第三方开源库，目前只支持使用 SMTP 服务器发送电子邮件，但是其 API 较为灵活，如果有其它的定制需求也可以轻易地借助其实现，这恰恰好符合我们的需求，因为目前我们只需要一个小而美的发送电子邮件的库就可以了。

#### 2.9.2.2.2 邮件工具库

在项目目录 `pkg` 下新建 email 目录并创建 email.go 文件，我们需要针对发送电子邮件的行为进行一些封装，写入如下代码：

```go
type Email struct {
	*SMTPInfo
}

type SMTPInfo struct {
	Host     string
	Port     int
	IsSSL    bool
	UserName string
	Password string
	From     string
}

func NewEmail(info *SMTPInfo) *Email {
	return &Email{SMTPInfo: info}
}

func (e *Email) SendMail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	dialer := gomail.NewDialer(e.Host, e.Port, e.UserName, e.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: e.IsSSL}
	return dialer.DialAndSend(m)
}
```

在上述代码中，我们定义了 SMTPInfo 结构体用于传递发送邮箱所必需的信息，而在 SendMail 方法中，我们首先调用 NewMessage 方法创建一个消息实例，可以用于设置邮件的一些必要信息，分别是：

- 发件人（From）
- 收件人（To）
- 邮件主题（Subject）
- 邮件正文（Body）

在完成消息实例的基本信息设置后，调用 NewDialer 方法创建一个新的 SMTP 拨号实例，设置对应的拨号信息用于连接 SMTP 服务器，最后再调用 DialAndSend 方法打开与 SMTP 服务器的连接并发送电子邮件。

#### 2.9.2.2.3 初始化配置信息

本次要做的发送电子邮件的行为，实际上你可以理解是与一个 SMTP 服务进行交互，那么除了自建 SMTP 服务器以外，我们可以使用目前市面上常见的邮件提供商，它们也是有提供 SMTP 服务的，首先我们打开项目的配置文件 config.yaml，新增如下 Email 的配置项：

```shell
Email:
  Host: smtp.qq.com
  Port: 465
  UserName: xxxx@qq.com
  Password: xxxxxxxx
  IsSSL: true
  From: xxxx@qq.com
  To:
    - xxxx@qq.com
```

通过 HOST 我们可以知道我用的是 QQ 邮件的 SMTP，这个只需要在”QQ 邮箱-设置-账户-POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV 服务“选项中将”POP3/SMTP 服务“和”IMAP/SMTP 服务“开启，然后根据所获取的 SMTP 账户密码进行设置即可，另外 SSL 是默认开启的。

另外需要特别提醒的一点是，我们所填写的 SMTP Server 的 HOST 端口号是 465，而常用的另外一类还有 25 端口号 ，但我强烈不建议使用 25，你应当切换为 465，因为 25 端口号在云服务厂商上是一个经常被默认封禁的端口号，并且不可解封，使用 25 端口，你很有可能会遇到部署进云服务环境后告警邮件无法正常发送出去的问题。

接下来我们在项目目录 `pkg/setting` 的 section.go 文件中，新增对应的 Email 配置项，如下：

```shell
type EmailSettingS struct {
	Host     string
	Port     int
	UserName string
	Password string
	IsSSL    bool
	From     string
	To       []string
}
```

并在在项目目录 `global` 的 setting.go 文件中，新增 Email 对应的配置全局对象，如下：

```go
var (
	...
	EmailSetting    *setting.EmailSettingS
	...
)
```

最后就是在项目根目录的 main.go 文件的 setupSetting 方法中，新增 Email 配置项的读取和映射，如下：

```go
func setupSetting() error {
	...
	err = s.ReadSection("Email", &global.EmailSetting)
	if err != nil {
		return err
	}
	...
}
```

### 2.9.2.3 编写中间件

我们打开项目目录 `internal/middleware` 并创建 recovery.go 文件，写入如下代码：

```go
func Recovery() gin.HandlerFunc {
	defailtMailer := email.NewEmail(&email.SMTPInfo{
		Host:     global.EmailSetting.Host,
		Port:     global.EmailSetting.Port,
		IsSSL:    global.EmailSetting.IsSSL,
		UserName: global.EmailSetting.UserName,
		Password: global.EmailSetting.Password,
		From:     global.EmailSetting.From,
	})
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				global.Logger.WithCallersFrames().Errorf("panic recover err: %v", err)

				err := defailtMailer.SendMail(
					global.EmailSetting.To,
					fmt.Sprintf("异常抛出，发生时间: %d", time.Now().Unix()),
					fmt.Sprintf("错误信息: %v", err),
				)
				if err != nil {
					global.Logger.Panicf("mail.SendMail err: %v", err)
				}

				app.NewResponse(c).ToErrorResponse(errcode.ServerError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
```

在本项目中，我们的 Mailer 是固定的，因此我们直接将其定义为了 defailtMailer，接着在捕获到异常后调用 SendMail 方法进行预警邮件发送，效果如下：

![image](https://golang2.eddycjy.com/images/ch2/panic_mail.jpg)

这里具体的邮件模板你可以根据实际情况进行定制。

## 2.9.3 服务信息存储

平时我们经常会需要在进程内上下文设置一些内部信息，例如是应用名称和应用版本号这类基本信息，也可以是业务属性的信息存储，例如是根据不同的租户号获取不同的数据库实例对象，这时候就需要有一个统一的地方处理，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-app-info.jpg)

我们打开项目下的 `internal/middleware` 目录并新建 app_info.go 文件，写入如下代码：

```go
func AppInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("app_name", "blog-service")
		c.Set("app_version", "1.0.0")
		c.Next()
	}
}
```

在上述代码中我们就需要用到 gin.Context 所提供的 setter 和 getter，在 gin 中称为元数据管理（Metadata Management），大致如下：

```go
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

func (c *Context) MustGet(key string) interface{} {...}
func (c *Context) GetString(key string) (s string) {...}
func (c *Context) GetBool(key string) (b bool) {...}
func (c *Context) GetInt(key string) (i int) {...}
func (c *Context) GetInt64(key string) (i64 int64) {...}
func (c *Context) GetFloat64(key string) (f64 float64) {...}
func (c *Context) GetTime(key string) (t time.Time) {...}
func (c *Context) GetDuration(key string) (d time.Duration) {...}
func (c *Context) GetStringSlice(key string) (ss []string) {...}
func (c *Context) GetStringMap(key string) (sm map[string]interface{}) {...}
func (c *Context) GetStringMapString(key string) (sms map[string]string) {...}
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {...}
```

实际上我们可以看到在 gin 中的 metadata，其实就是利用内部实现的 gin.Context 中的 Keys 进行存储的，并配套了多种类型的获取和设置方法，相当的方便。另外我们可以注意到在默认的 Get 和 Set 方法中，传入和返回的都是 interface 类型，实际在业务属性的初始化逻辑处理中，我们可以通过对返回的 interface 进行类型断言，就可以获取到我们所需要的类型了。

## 2.9.4 接口限流控制

在应用程序的运行过程中，会不断地有新的客户端进行访问，而有时候会突然出现流量高峰（例如：营销活动），如果不及时进行削峰，资源整体又跟不上，那就很有可能会造成事故，因此我们常常会才有多种手段进行限流削峰，而针对应用接口进行限流控制就是其中一种方法，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-limiter.jpg)

### 2.9.4.1 安装

```shell
$ go get -u github.com/juju/ratelimit@v1.0.1
```

Ratelimit 提供了一个简单又高效的令牌桶实现，能够提供大量的方法帮助我们实现限流器的逻辑。

### 2.9.4.2 限流控制

#### 2.9.4.2.1 LimiterIface

我们打开项目的 `pkg/limiter` 目录并新建 limiter.go 文件，写入如下代码：

```go
type LimiterIface interface {
	Key(c *gin.Context) string
	GetBucket(key string) (*ratelimit.Bucket, bool)
	AddBuckets(rules ...LimiterBucketRule) LimiterIface
}

type Limiter struct {
	limiterBuckets map[string]*ratelimit.Bucket
}

type LimiterBucketRule struct {
	Key          string
	FillInterval time.Duration
	Capacity     int64
	Quantum      int64
}
```

在上述代码中，我们声明了 LimiterIface 接口，用于定义当前限流器所必须要的方法。

为什么要这么做呢，实际上需要知道一点，限流器是存在多种实现的，可能某一类接口需要限流器 A，另外一类接口需要限流器 B，所采用的策略不是完全一致的，因此我们需要声明 LimiterIfac 这类通用接口，保证其接口的设计，我们初步的在 Iface 接口中，一共声明了三个方法，如下：

- Key：获取对应的限流器的键值对名称。
- GetBucket：获取令牌桶。
- AddBuckets：新增多个令牌桶。

同时我们定义 Limiter 结构体用于存储令牌桶与键值对名称的映射关系，并定义 LimiterBucketRule 结构体用于存储令牌桶的一些相应规则属性，如下：

- Key：自定义键值对名称。
- FillInterval：间隔多久时间放 N 个令牌。
- Capacity：令牌桶的容量。
- Quantum：每次到达间隔时间后所放的具体令牌数量。

至此我们就完成了一个 Limter 最基本的属性定义了，接下来我们将针对不同的情况实现我们这个项目中的限流器。

#### 2.9.4.2.2 MethodLimiter

我们第一个编写的简单限流器的主要功能是针对路由进行限流，因为在项目中，我们可能只需要对某一部分的接口进行流量调控，我们打开项目下的 `pkg/limiter` 目录并新建 method_limiter.go 文件，写入如下代码：

```go
type MethodLimiter struct {
	*Limiter
}

func NewMethodLimiter() LimiterIface {
	return MethodLimiter{
		Limiter: &Limiter{limiterBuckets: make(map[string]*ratelimit.Bucket)},
	}
}

func (l MethodLimiter) Key(c *gin.Context) string {
	uri := c.Request.RequestURI
	index := strings.Index(uri, "?")
	if index == -1 {
		return uri
	}

	return uri[:index]
}

func (l MethodLimiter) GetBucket(key string) (*ratelimit.Bucket, bool) {
	bucket, ok := l.limiterBuckets[key]
	return bucket, ok
}

func (l MethodLimiter) AddBuckets(rules ...LimiterBucketRule) LimiterIface {
	for _, rule := range rules {
		if _, ok := l.limiterBuckets[rule.Key]; !ok {
			l.limiterBuckets[rule.Key] = ratelimit.NewBucketWithQuantum(rule.FillInterval, rule.Capacity, rule.Quantum)
		}
	}

	return l
}
```

在上述代码中，我们针对 LimiterIface 接口实现了我们的 MethodLimiter 限流器，主要逻辑是在 Key 方法中根据 RequestURI 切割出核心路由作为键值对名称，并在 GetBucket 和 AddBuckets 进行获取和设置 Bucket 的对应逻辑。

### 2.9.4.3 编写中间件

在完成了限流器的逻辑编写后，打开项目下的 `internal/middleware` 目录并新建 limiter.go 文件，将整体的限流器与对应的中间件逻辑串联起来，写入如下代码：

```go
func RateLimiter(l limiter.LimiterIface) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := l.Key(c)
		if bucket, ok := l.GetBucket(key); ok {
			count := bucket.TakeAvailable(1)
			if count == 0 {
				response := app.NewResponse(c)
				response.ToErrorResponse(errcode.TooManyRequests)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
```

在 RateLimiter 中间件中，需要注意的是入参应该为 LimiterIface 接口类型，这样子的话只要符合该接口类型的具体限流器实现都可以传入并使用，另外比较重要的就是 TakeAvailable 方法，它会占用存储桶中立即可用的令牌的数量，返回值为删除的令牌数，如果没有可用的令牌，将会返回 0，也就是已经超出配额了，因此这时候我们将返回 errcode.TooManyRequest 状态告诉客户端需要减缓并控制请求速度。

## 2.9.5 统一超时控制

在应用程序的运行中，常常会遇到一个头疼的问题，调用链如果是应用 A =》应用 B =》应用 C，那如果应用 C 出现了问题，在没有任何约束的情况下持续调用，就会导致应用 A、B、C 均出现问题，也就是很常见的上下游应用的互相影响，导致连环反应，最终使得整个集群应用出现一定规模的不可用，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-context-boom.jpg)

为了规避这种情况，最简单也是最基本的一个约束点，那就是统一的在应用程序中针对所有请求都进行一个最基本的超时时间控制，如下图：

![image](https://golang2.eddycjy.com/images/ch2/app-context-deadline.jpg)

为此我们就编写一个上下文超时时间控制的中间件来实现这个需求，打开项目下的 `internal/middleware` 目录并新建 context_timeout.go 文件，如下：

```go
func ContextTimeout(t time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), t)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
```

在上述代码中，我们调用了 context.WithTimeout 方法设置当前 context 的超时时间，并重新赋予给了 gin.Context，这样子在当前请求运行到指定的时间后，在使用了该 context 的运行流程就会针对 context 所提供的超时时间进行处理，并在指定的时间进行取消行为。效果如下：

```go
_, err := ctxhttp.Get(c.Request.Context(), http.DefaultClient, "https://www.google.com/")
if err != nil {
    log.Fatalf("ctxhttp.Get err: %v", err)
}
```

我们需要将我们设置了超时的 `c.Request.Context()` 给传递进去，在验证时你可以将默认超时时间调短来进行调试，其最终输出结果：

```shell
ctxhttp.Get err: context deadline exceeded
exit status 1
```

最后由于已经到达了截止时间，因此返回 `context deadline exceeded` 错误提示信息。另外这里还需要注意，如果你在进行多应用/服务的调用时，把父级的上下文信息（ctx）不断地传递下去，那么在统计超时控制的中间件中所设置的超时时间，其实是针对整条链路的，而不是针对单单每一条，如果你需要针对额外的链路进行超时时间的调整，那么只需要调用像 `context.WithTimeout` 等方法对父级 ctx 进行设置，然后取得子级 ctx，再进行新的上下文传递就可以了。

## 2.9.6 注册中间件

在完成一连串的通用中间件编写后，打开项目目录 `internal/routers` 下的 router.go 文件，修改注册应用中间件的逻辑，如下：

```go
var methodLimiters = limiter.NewMethodLimiter().AddBuckets(limiter.LimiterBucketRule{
	Key:          "/auth",
	FillInterval: time.Second,
	Capacity:     10,
	Quantum:      10,
})

func NewRouter() *gin.Engine {
	r := gin.New()
	if global.ServerSetting.RunMode == "debug" {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	} else {
		r.Use(middleware.AccessLog())
		r.Use(middleware.Recovery())
	}

	r.Use(middleware.RateLimiter(methodLimiters))
	r.Use(middleware.ContextTimeout(60 * time.Second))
	r.Use(middleware.Translations())
	...
	apiv1.Use(middleware.JWT()){...}
```

在上述代码中，我们根据不同的部署环境（RunMode）进行了应用中间件的设置，因为实际上在使用了自定义的 Logger 和 Recovery 后，就没有必要使用 gin 原有所提供的了，而在本地开发环境中，可能没有齐全应用生态圈，因此需要进行特殊处理。另外在常规项目中，自定义的中间件不仅包含了基本的功能，还包含了很多定制化的功能，同时在注册顺序上也注意，Recovery 这类应用中间件应当尽可能的早注册，这根据实际所要应用的中间件情况进行顺序定制就可以了。

这里我们可以看到 `middleware.ContextTimeout` 是写死的 60 秒，在此交给你一个小任务，你可以对其进行配置化（映射配置和秒数初始化），将超时的时间配置调整到配置文件中，而不是在代码中硬编码，最终结果应当如下：

```go
r.Use(middleware.ContextTimeout(global.AppSetting.DefaultContextTimeout))
```

这样子的话，以后修改超时的时间就只需要通过修改配置文件就可以解决了，不需要人为的修改代码，甚至可以不需要开发人员的直接参与，让运维同事确认后直接修改即可。

