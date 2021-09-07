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

