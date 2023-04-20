# Micro Service Demo

## 1. 项目简介

演示代码是一个简单的图片处理服务

```text
            +---------+                   +---------+
            | Gateway +------------------>|  Webdav |
            +----+----+                   +---------+
                 |                             ^
                 |        +----------+         |
                 +------->| Process  +---------+
                          +----------+
```

- Gateway 应用入口
- Webdav 负责照片数据的 CRUD
- Process 是图片处理，包含多种处理算法：灰化、加水印等

Gateway 可以直接调用 Webdav 接口，查询图片列表；然后调用 Process 接口（Process 从 Webdav 拿到图片，然后图片处理之后再返回给 Gateway）展示图片。

# 2. 项目部署

## 2.1 环境准备

1. AIO 环境，4Core / 8G / 40G，CentOS 7.9
2. 部署 KubeClipper 1.3.2 + K8S v1.23.6，参考 [Github](https://github.com/wu-wenxiang/lab-kubernetes/blob/master/doc/cloudnative-and-mircoservice.md#322-%E5%AE%89%E8%A3%85-k8s-1236) 或 [Gitee](https://gitee.com/wu-wen-xiang/lab-kubernetes/blob/master/doc/cloudnative-and-mircoservice.md#322-%E5%AE%89%E8%A3%85-k8s-1236)
3. 配置默认的、支持动态分配存储的 Storage Class，参考 [Github](https://github.com/wu-wenxiang/lab-kubernetes/blob/master/doc/kubernetes-best-practices.md#45-local-%E5%92%8C%E5%8A%A8%E6%80%81%E5%88%86%E9%85%8D) 或 [Gitee](https://gitee.com/wu-wen-xiang/lab-kubernetes/blob/master/doc/kubernetes-best-practices.md#45-local-%E5%92%8C%E5%8A%A8%E6%80%81%E5%88%86%E9%85%8D)

## 2.2 部署项目到标准 K8S 环境

项目可以通过 [deploy.yaml 文件](manifest/deploy.yaml) 部署到标准 K8S 环境中。

```bash
# 清除 namespace ms-demo
kubectl delete ns ms-demo

# 创建 namespace ms-demo
kubectl create ns ms-demo

# 部署项目到 ms-demo namespace
wget https://gitee.com/dev-99cloud/micro-service-demo/raw/master/manifest/deploy.yaml
kubectl -n ms-demo apply -f deploy.yaml
```

然后访问 `http://<IP>:30086`，可以看到 Gateway 页面，可以 CRUD 图片。

## 2.3 图片处理算法展示

调整 deploy.yaml 中，令：

1. `GRAYSCALE = "true"`，重新 apply，可以看到图片的灰化效果
1. 配置 `WATERMARK = "hello"`，可以看到水印效果。

## 2.4 模拟访问失败的情况

访问 `/process/_statusCode_`，可以返回 500

```console
# curl -i http://47.242.127.16:30086/process/_statusCode_
HTTP/1.1 500 Internal Server Error
Pod-Name: gateway-7f745b5c5d-sbn8h
Date: Sun, 23 Oct 2022 10:05:07 GMT
Content-Length: 21
Content-Type: text/plain; charset=utf-8

Internal Server Error
```
