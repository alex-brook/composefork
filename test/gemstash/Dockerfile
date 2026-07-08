FROM docker.io/ruby:3.4.10-slim-trixie
RUN gem install gemstash & mkdir -p /root/.gemstash
COPY config.yml ~/root/.gemstash/config.yml
