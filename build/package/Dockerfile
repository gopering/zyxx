FROM registry.cn-hangzhou.aliyuncs.com/louis-mirror/alpine:1.1
LABEL maintainer="zhangzhen<zhangzhen@simetra.cn>"

USER root

STOPSIGNAL 37

# 设置时区为上海
ENV TZ=Asia/Shanghai

WORKDIR /usr/local/zyxx

# 拷贝脚本
# 健康检查配置: sh /root/healthcheck.sh
# prestop配置: sh /root/run.sh prestop
COPY ./scripts/*.sh  /root/

# 增加执行权限
RUN chmod +x /root/*.sh

# 拷贝程序文件
COPY ./release/  ./

COPY ./scripts/* /etc/kickStart.d/

# COPY ./resource ./

# 增加执行权限
RUN chmod +x /etc/kickStart.d/*.sh

ENTRYPOINT /etc/kickStart.d/dockerfile-entry.sh
