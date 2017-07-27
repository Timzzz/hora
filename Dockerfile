FROM nginx
ADD config.toml /config.toml
ADD hora /hora
ADD VERSION /VERSION

EXPOSE 8080
CMD ["/hora"]
