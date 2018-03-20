FROM alpine:3.6 as ffmpeg-builder

RUN apk add --no-cache \
  coreutils \
  openssl \
  bash \
  build-base \
  autoconf \
  automake \
  libtool \
  git \
  yasm \
  zlib-dev \
  openssl-dev \
  lame-dev \
  libogg-dev \
  libvpx-dev \
  x265-dev

# some -dev alpine packages lack .a files in 3.6 (some fixed in edge)
RUN \
  FDK_AAC_VERSION=0.1.5 && \
  wget -O - https://github.com/mstorsjo/fdk-aac/archive/v$FDK_AAC_VERSION.tar.gz | tar xz && \
  cd fdk-aac-$FDK_AAC_VERSION && \
  ./autogen.sh && \
  ./configure --enable-static && \
  make -j4 install

RUN \
  VORBIS_VERSION=1.3.5 && \
  wget -O - https://downloads.xiph.org/releases/vorbis/libvorbis-$VORBIS_VERSION.tar.gz | tar xz && \
  cd libvorbis-$VORBIS_VERSION && \
  CFLAGS="-fno-strict-overflow -fstack-protector-all -fPIE" LDFLAGS="-Wl,-z,relro -Wl,-z,now -fPIE -pie" \
  ./configure --enable-static && \
  make -j4 install

RUN \
  OPUS_VERSION=1.2.1 && \
  wget -O - https://archive.mozilla.org/pub/opus/opus-$OPUS_VERSION.tar.gz | tar xz && \
  cd opus-$OPUS_VERSION && \
  CFLAGS="-fno-strict-overflow -fstack-protector-all -fPIE" LDFLAGS="-Wl,-z,relro -Wl,-z,now -fPIE -pie" \
  ./configure --enable-static && \
  make -j4 install

# require libogg to build
RUN \
  THEORA_VERSION=1.1.1 && \
  wget -O - https://downloads.xiph.org/releases/theora/libtheora-$THEORA_VERSION.tar.bz2 | tar xj && \
  cd libtheora-$THEORA_VERSION && \
  CFLAGS="-fno-strict-overflow -fstack-protector-all -fPIE" LDFLAGS="-Wl,-z,relro -Wl,-z,now -fPIE -pie" \
  ./configure --enable-pic --enable-static && \
  make -j4 install

# x264 only has a "stable" branch no tags
RUN \
  X264_VERSION=aaa9aa83a111ed6f1db253d5afa91c5fc844583f && \
  git clone git://git.videolan.org/x264.git && \
  cd x264 && \
  git checkout $X264_VERSION && \
  CFLAGS="-fno-strict-overflow -fstack-protector-all -fPIE" LDFLAGS="-Wl,-z,relro -Wl,-z,now -fPIE -pie" \
  ./configure --enable-pic --enable-static && make -j4 install

# note that this will produce a "static" PIE binary with no dynamic lib deps
ENV FFMPEG_VERSION=n3.4.2
RUN \
  git clone --branch $FFMPEG_VERSION --depth 1 https://github.com/FFmpeg/FFmpeg.git && \
  cd FFmpeg && \
  ./configure \
  --toolchain=hardened \
  --disable-shared \
  --enable-static \
  --pkg-config-flags=--static \
  --extra-ldflags=-static \
  --extra-cflags=-static \
  --enable-gpl \
  --enable-nonfree \
  --enable-openssl \
  --disable-ffserver \
  --disable-doc \
  --disable-ffplay \
  --enable-libmp3lame \
  --enable-libfdk-aac \
  --enable-libvorbis \
  --enable-libopus \
  --enable-libtheora \
  --enable-libvpx \
  --enable-libx264 \
  --enable-libx265 \
  && \
  make -j4 install && \
  ldd /usr/local/bin/ffmpeg | grep -vq lib && \
  ldd /usr/local/bin/ffprobe | grep -vq lib

FROM golang:1.10-stretch as ydls-builder
ENV YDL_VERSION=2018.03.20
ENV CONFIG=/etc/ydls.json

RUN \
  curl -L -o /usr/local/bin/youtube-dl https://yt-dl.org/downloads/$YDL_VERSION/youtube-dl && \
  chmod a+x /usr/local/bin/youtube-dl
COPY --from=ffmpeg-builder \
  /usr/local/bin/ffmpeg \
  /usr/local/bin/ffprobe \
  /usr/local/bin/

COPY . /go/src/github.com/wader/ydls/
COPY ydls.json /etc

WORKDIR /go/src/github.com/wader/ydls

RUN TEST_FFMPEG=1 TEST_YOUTUBEDL=1 TEST_NETWORK=1 go test -v -cover -race ./...
RUN go install -installsuffix netgo -tags netgo -ldflags "-X main.gitCommit=$(git describe --always)" ./cmd/ydls
RUN \
  ldd /go/bin/ydls | grep -q "not a dynamic executable" && \
  cmd/ydls/ydls_server_test.sh && \
  cmd/ydls/ydls_get_test.sh

FROM alpine:3.6
LABEL maintainer="Mattias Wadman mattias.wadman@gmail.com"
ENV PORT=8080
ENV LISTEN=:$PORT
ENV CONFIG=/etc/ydls.json

RUN apk add --no-cache \
  ca-certificates \
  tini \
  python \
  rtmpdump \
  mplayer
COPY --from=ffmpeg-builder \
  /usr/local/bin/ffmpeg \
  /usr/local/bin/ffprobe \
  /usr/local/bin/
COPY --from=ydls-builder \
  /go/bin/ydls \
  /usr/local/bin/youtube-dl \
  /usr/local/bin/
COPY entrypoint.sh /usr/local/bin
COPY ydls.json /etc

# make sure all binaries work and do some sanity checks (https, DNS)
RUN \
  youtube-dl --version && \
  ffmpeg -version && \
  ffprobe -version && \
  ydls -version && \
  ffmpeg -i https://www.google.com 2>&1 | grep -q "Invalid data found when processing input"

USER nobody
EXPOSE $PORT/tcp
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
