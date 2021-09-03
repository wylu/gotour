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

