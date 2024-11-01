# 定义Go编译器和构建标志  
GOCMD=go  
GOBUILD=$(GOCMD) build -o myapp . 

# 定义Docker相关变量 
IMAGE_NAME=my-go-app  
DOCKERFILE=Dockerfile 

# 编译Go项目 
build:  
	$(GOBUILD)  

# 清理构建的文件（可选）
clean:  
	$(GOCMD) clean -i -cache -modcache  
	rm -f myapp  

# 构建Docker镜像
image:  
	docker build -t $(IMAGE_NAME) -f $(DOCKERFILE) . 

# 运行Docker容器（仅用于测试，可选）  
run:  
	docker run --rm -it -p 8080:8080 $(IMAGE_NAME)  

# 完整的构建和打包流程（包含编译和打包Docker镜像）  
all: clean build image
