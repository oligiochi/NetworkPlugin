FROM alpine

ENV LD_LIBRARY_PATH=/usr/local/lib/

RUN apk add -U --no-cache \
    linux-headers \
    gcc \
    libc-dev \
    clang \
    llvm \
    git \
    automake \
    autoconf \
    libtool \
    make \
    cmake \
    musl-dev \
    patch

WORKDIR /tmp

RUN git clone https://github.com/rd235/s2argv-execs.git && cd s2argv-execs/ && \
    autoreconf -if && ./configure && make && make install && \
    cd .. && rm -rf s2argv-execs/

COPY support-for-jumbo.patch /tmp
RUN git clone https://github.com/virtualsquare/vde-2 && \
    mv support-for-jumbo.patch vde-2/support-for-jumbo.patch && \
    cd vde-2/ && patch -p1 < support-for-jumbo.patch && \
    mkdir build && cd build && \
    cmake -DENABLE_CRYPTCAB=OFF -DENABLE_VDE_OVER_NS=OFF -DENABLE_ROUTER=OFF -DENABLE_VXLAN=OFF -DENABLE_TUNTAP=OFF -DENABLE_PCAP=OFF .. && \
    make && make install && \
    cd .. && rm -rf vde-2/

RUN git clone https://github.com/rd235/vdeplug4.git && cd vdeplug4/ && \
    sed -i 's/ | RTLD_DEEPBIND//' libvdeplug4/libvdeplug.c && \
    sed -i '1i#include <libgen.h>' srcvdeplug4/vde_plug.c && \
    mkdir build && cd build && cmake .. && make && make install && \
    cd ../.. && rm -rf vdeplug4/

RUN apk del -U --no-cache \
    linux-headers \
    gcc \
    libc-dev \
    clang \
    llvm \
    git \
    automake \
    autoconf \
    libtool \
    make \
    cmake \
    musl-dev \
    patch

WORKDIR /app

COPY entrypoint.sh entrypoint.sh
COPY katharanp katharanp
COPY vde_ext /usr/local/bin/vde_ext
RUN chmod +x entrypoint.sh katharanp /usr/local/bin/vde_ext
