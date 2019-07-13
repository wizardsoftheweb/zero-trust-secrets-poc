FROM        ubuntu:18.04

MAINTAINER  CJ Harries <cj@wizardsoftheweb.pro>

ENV         UID '9001'
ENV         GID '9001'
ENV         USER 'appuser'
ENV         GROUP "$USER"
ENV         HOME_DIR "/$USER"
ENV         SHELL "/bin/bash"

RUN         apt-get update -q=2  && \
            apt-get upgrade -q=2 && \
            apt-get install -q=2 --no-install-recommends  \
                gnupg && \
            mkdir -p "$HOME_DIR" && \
            groupadd -g "$GID" "$GROUP" && \
            useradd -g "$GID" -u "$UID" -d "$HOME_DIR" -s "$SHELL" "$USER" && \
            chown -R "${USER}:${GROUP}" "$HOME_DIR"

#USER        "$USER"

#RUN         rm -rf /var/lib/apt/lists/* && \
#            apt-get purge && \
#            apt-get clean
CMD         ["/bin/bash", "-c", "while true; do sleep 1; done"]
