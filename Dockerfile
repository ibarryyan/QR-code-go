# 使用最新的Ubuntu作为基础镜像
FROM golang:latest

# 设置工作目录
WORKDIR /app

# 将Go代码复制到容器中
COPY . /app

# 编译Go代码
RUN go env -w GOPROXY=https://goproxy.io
RUN go mod tidy
RUN go build -o myapp

# 设置容器启动时执行的命令
ENTRYPOINT ["/app/myapp"]