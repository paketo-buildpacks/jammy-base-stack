FROM ubuntu:jammy

ARG sources
ARG packages

RUN echo "$sources" > /etc/apt/sources.list

RUN echo "debconf debconf/frontend select noninteractive" | debconf-set-selections && \
  export DEBIAN_FRONTEND=noninteractive && \
  apt-get -y $package_args update && \
  apt-get -y $package_args upgrade && \
  apt-get -y $package_args install locales && \
  locale-gen en_US.UTF-8 && \
  update-locale LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8 && \
  apt-get -y $package_args install $packages && \
  rm -rf /var/lib/apt/lists/* /tmp/*

RUN rm /etc/os-release && cat /usr/lib/os-release | \
    sed -e 's#HOME_URL=.*#HOME_URL="https://github.com/paketo-buildpacks/jammy-base-stack"#' \
        -e 's#SUPPORT_URL=.*#SUPPORT_URL="https://github.com/paketo-buildpacks/jammy-base-stack/blob/main/README.md"#' \
        -e 's#BUG_REPORT_URL=.*#BUG_REPORT_URL="https://github.com/paketo-buildpacks/jammy-base-stack/issues/new"#' \
  > /etc/os-release \
  && echo 'PRETTY_NAME="Paketo Buildpacks Base Jammy"' >> /etc/os-release
