**Update Docs**

```shell
# 1.16 及以上版本
$ go install github.com/swaggo/swag/cmd/swag@latest

$ swag init -g router/router.go

$ go run main.go
```

访问：localhost:8880/swagger/index.html



---

[Gin Docs](https://gin-gonic.com/docs/)

[Gorm Docs](https://gorm.io/zh_CN/docs/)

[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)


---

**攻击算法管理**

攻击算法

Parameters:

- 用户上传的模型 - 我是不是直接传递个文件路径就好了

- 数据集 - 这个怎么办：用户自己编写dataloader函数？