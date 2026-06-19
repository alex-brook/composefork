FROM ruby:4-trixie

WORKDIR /rails

COPY .devcontainer/entrypoint.sh /usr/bin/entrypoint
RUN chmod +x /usr/bin/entrypoint

ENTRYPOINT ["entrypoint"]
EXPOSE 3000
# CMD ["bin/dev"]
CMD ["sleep", "infinity"]
