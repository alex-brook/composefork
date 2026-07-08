FROM docker.io/ruby:3.4.10-trixie
RUN apt update -yq \
  && apt install -yq build-essential tini \
  && gem install gemstash \
  && mkdir -p /root/.gemstash
COPY config.yml ~/root/.gemstash/config.yml
ENTRYPOINT ["tini", "--"]
CMD ["gem", "exec", "gemstash", "start"]
