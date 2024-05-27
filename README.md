# 实现简单Docker

## 主要实现功能
//TODO


## 整体流程图
说明： 该流程图只包含最**核心**的容器创建，其他诸如 log、ps、commit等命令代码上实现了但并未在流程图上体现
    1. 容器的创建（Namespace隔离、cgroup 资源限制、rootfs文件系统）
    2. 网络相关（创建虚拟网桥、IP分配、设置SNAT和DNAT）
流程图如下：
![Docker项目梳理](https://github.com/Chen-Litao/My_Docker/assets/138087482/a348c1ba-a16f-431a-ba53-a0aa1c5ff1bd)
