---
name: go-naming-conventions
description: 飞鸟快验Go项目中文命名规范指南。此技能应在编写或修改本项目Go代码时始终激活，确保AI严格遵守项目独特的中文命名规范
---

# 项目分层设计
本系统的后端实现并非简单的“控制器直连数据库”，而是形成了较明确的分层：
## Router 层
路径`new/app/router`
实现 URL 绑定、路由分域和接口分发,中间件等。
## Controller 层
路径`new/app/controller`
负责参数接收、请求编排和响应格式统一,简单单一功能直接调用Service,不可处理复杂功能和需要多表联动功能。
## Logic 层
路径`new/app/logic`
负责核心业务运算，逻辑处理,以及多表事务等复杂功能实现。只操作单表调用Service增删改查,等简单功能不可放到这里.
## Service 层
路径`new/app/service`
负责数据库对象级操作及通用增删改查封装。
## Models 层
路径`new/app/models`
负责数据库结构体、通用请求响应体、通用常量定义。

# 飞鸟快验 Go 项目中文命名规范

## 目的
确保所有新生成的Go代码严格遵守本项目的中文命名体系
旧代码暂时不用刻意重构风格,仅新代码遵守即可
## 核心规则
### 1. 变量命名：必须使用作用域前缀 + 中文

局部变量必须用 `局_` 前缀，包级私有变量用 `集_` 前缀,跨包导出全局变量 使用 `Q全_`前缀

| 作用域     | 前缀格式 | 正确示例                         | 错误示例               |
|---------|----------|------------------------------|--------------------|
| 局部变量    | `局_` | `局_计数`, `局_子级代理id数组`, `局_返回` | `计数`, `局计数`        |
| 包级变量    | `集_` | `集_运行目录`, `集_userAPi路由强制RSA` | `集运行目录`, `全局_运行目录` |
| 全局变量 | `Q全_` | `Q全_缓存`, `Q全_快验`, `Q全_系统信息`  | `缓存`, `快验`, `系统信息` |

### 2. 函数命名：中文函数名 + 分类前缀

需要导出函数名使用  {分类拼音首字母大写}{中文分类名}_{中文名称}，例如 Z正则_校验密码、K卡类_详情。
不需要导出函数名使用   {中文分类名}_{尽量中文函数名}
参数名无需前缀,中英混合也可以,易理解即可

| 前缀    | 含义 |是否导出| 示例 |
|-------|------|------|------|
| `Z正则` | 正则/校验 |是| `Z正则_校验密码`, `Z正则_校验用户名` |
| `包` | 包名 | 否|`包_包名结构体方法`  |
| `代理` | 代理 | 否| `代理_取Id代理级别`  |
| `卡类` | 卡类 | 否|  `卡类_可制卡类树形框列表` |


### 3. 结构体命名

#### 3a. 业务结构体命名：使用分类前缀，不使用变量作用域前缀

需要导出结构体使用  {分类拼音首字母大写}{中文分类名}_{尽量中文结构体名}
不需要导出结构体使用   {中文分类名}_{尽量中文结构体名}
无分类可以不要分类名  

```go
// ✅ 正确
type 配置_代理基础设置 struct {}
type 结构_AppIdNameList struct {}
type 卡号_单卡号 struct {}
type 结构_登录请求 struct {}
type K卡类_详情 struct {}   

// ❌ 错误
type 局_代理可制卡类授权 struct {}   // 结构体不需要"局_"前缀
type 代理可制卡类授权 struct {}    // 没有前缀
```
#### 3b. 英文结构体名
仅限中文不方便的地方使用
需要导出结构体使用 大驼峰
不需要导出结构体使用  小驼峰
#### 3c. 数据库表结构

数据库表结构体使用 `DB_`  前缀 +小驼峰
存放路径`new/app/models/db`
```go
// ✅ 正确
type DB_user struct {}   
type DB_admin struct {}
type DB_agentLog struct {}
type DB_agentClass struct {}
type DB_agentUser struct {}  //DB_ 后面的表对象名使用小驼峰，如 DB_agentUser

// ❌ 错误
type db_user struct {}        // db前缀应大写为DB
```

数据库表名 使用全部小写 下划线命名法  `db_` 前缀
```text 
// ✅ 正确
`db_check_in_task_log`
`db_tong_ji_zai_xian`

// ❌ 错误
`db_Log_UserActive`   // 没有全部小写 
```



#### 3d. 请求/响应结构体命名模式

请求结构体用 `请求_` 前缀，响应结构体用 `响应_` 前缀：

```go
// ✅ 正确
type 请求_create struct {}
type 请求_agentInventoryGetInfo struct {}
type 响应_linkUserGetList struct {}
type 请求_登录 struct {}
type 响应_登录 struct {}

// ❌ 错误
type Request_Create struct {}    // 不用英文Request，用中文"请求"
type 响应create struct {}        // 没有使用前缀
```

### 4. 结构体字段命名

#### 4a. 中文字段名：直接中文，不加前缀
 成员名需要全部导出,首字拼音首拼大写
 中文字段本身没有大小写，必须加拼音首字母大写前缀才能导出，例如 B绑定信息
```go
// ✅ 正确 
type K快验_帐号信息 struct {
    B绑定信息      string
    Y用户类型      string
    D到期时间      int64
    Z注册时间      int
    D登录时间      int
    D登录IP      string
    Y余额        float64
    J积分        float64
    H会员帐号      string
    H会员密码      string
    Y用户备注      string
}

// ❌ 错误
type K快验帐号信息 struct {
    绑定信息      string    // 缺少首字拼音首拼大写
    局_绑定信息    string    // 字段不需要"局_"前缀
    b绑定信息      string    // 首字拼音首拼大写 没有大写,导致无法导出
}
```
#### 4b. 英文字段名：遵循Go惯例，首字母大写用于导出
`json` tag 使用小驼峰；`gorm:"column:..."` 按数据库字段实际名称填写。

```go
// ✅ 正确（DB_User结构体中的英文字段）
type DB_user struct {
    Id                  int     `json:"id" gorm:"column:id;primarykey"`
    User                string  `json:"user" gorm:"column:user;index"`
    Status              int     `json:"status" gorm:"column:status;default:1"`
    Rmb                 float64 `json:"rmb" gorm:"column:rmb;type:decimal(10,2)"`
    UpAgentId           int     `json:"upAgentId" gorm:"column:upAgentId"`
    LoginTime           int64   `json:"loginTime" gorm:"column:loginTime"`
}

// ❌ 错误
type DB_user struct {
    id                  int     // 未导出，其他包无法访问
    user                string  // 未导出
    UpAgentid           int     // id应大写为Id
    LoginTime           int   `json:"LoginTime"`  // json tag 应使用小驼峰 loginTime
}
```

### 5. 方法接收者命名
接收者变量名 统一使用 `j`

### 6. 方法命名
需要导出方法名使用  {首字拼音首拼大写}{中文函数名}
不需要导出方法名使用    {尽量中文函数名}
参数名无需前缀,中英混合也可以,易理解即可


| 方法名       |是否导出| 含义               |  
|-----------|------|------------------| 
| `C初始化`    |是| 初始化              | 
| `Q取错误信息`  | 是| 获取错误信息           | 
| `Q请求基本信息` | 是| 执行功能             |  
| `Info` | 是| 执行功能             |  
| `Create` | 是| 执行功能             |  
| `算法md5`   | 否| 内部算法名称,无需外部调用不导出 | 
```go
// ✅ 正确
func (j *AgentUserFull) Info(c *gin.Context) {}
func (j *UserClass) Create(info DB.DB_UserConfig) (row int64, err error) {}
func (j *Sdk) Q取错误信息()(错误信息 string) {}
func (j *Sdk) C初始化(token string)(err error) {}

// ❌ 错误
func (self *AgentUserFull) Info(c *gin.Context) {}   // 不要用self/this
func (agentUserFull *AgentUserFull) Info(c *gin.Context) {}  //   接收者变量名没有使用`j`
```
 
---

# 导入库约定

当代码中需要使用以下包时，必须按下列形式导入，包括点导入 `.` 和别名，不要改用默认包名或其他别名。

```go
import (
    . "EFunc/utils"
    . "server/app/models/request"
    . "server/app/models/response"
    dbm "server/app/models/db"
    utils2 "server/utils"
)
```



## 特殊注意事项
1. **匿名请求结构体**：项目中大量使用 `var 请求 struct {...}` 的模式来定义局部请求变量，这是项目惯例，应保持一致。
