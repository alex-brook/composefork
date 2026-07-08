FROM ruby:4-trixie

WORKDIR /rails

RUN apt update && apt install -y tini

COPY .devcontainer/entrypoint.sh /usr/bin/entrypoint
RUN chmod +x /usr/bin/entrypoint

ENTRYPOINT ["entrypoint"]
EXPOSE 3000
CMD ["tini", "bin/dev"]
