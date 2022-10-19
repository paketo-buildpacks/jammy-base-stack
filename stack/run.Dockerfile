FROM ubuntu:jammy

ARG sources
ARG packages
ARG package_args='--no-install-recommends'

RUN echo "$sources" > /etc/apt/sources.list
RUN echo "Package: $packages\nPin: release c=multiverse\nPin-Priority: -1\n\nPackage: $packages\nPin: release c=restricted\nPin-Priority: -1\n" > /etc/apt/preferences

RUN echo "debconf debconf/frontend select noninteractive" | debconf-set-selections && \
  export DEBIAN_FRONTEND=noninteractive && \
  apt-get -y $package_args update && \
  apt-get -y $package_args upgrade && \
  apt-get -y $package_args install locales && \
  locale-gen en_US.UTF-8 && \
  update-locale LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8 && \
  apt-get -y $package_args install $packages && \
  find /usr/share/doc/*/* ! -name copyright | xargs rm -rf && \
  rm -rf \
    /usr/share/man/* /usr/share/info/* \
    /usr/share/groff/* /usr/share/lintian/* /usr/share/linda/* \
    /var/lib/apt/lists/* /tmp/* /etc/apt/preferences

# remove /etc/os-release first as the test framework does not follow symlinks
RUN rm /etc/os-release && cat /usr/lib/os-release | \
    sed -e 's#PRETTY_NAME=.*#PRETTY_NAME="Paketo Buildpacks Base Jammy"#' \
        -e 's#HOME_URL=.*#HOME_URL="https://github.com/paketo-buildpacks/jammy-base-stack"#' \
        -e 's#SUPPORT_URL=.*#SUPPORT_URL="https://github.com/paketo-buildpacks/jammy-base-stack/blob/main/README.md"#' \
        -e 's#BUG_REPORT_URL=.*#BUG_REPORT_URL="https://github.com/paketo-buildpacks/jammy-base-stack/issues/new"#' \
  > /etc/os-release && rm /usr/lib/os-release
