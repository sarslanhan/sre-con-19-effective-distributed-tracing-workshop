FROM maven:slim

COPY pom.xml /tmp/pom.xml
RUN mvn -B -f /tmp/pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:resolve dependency:resolve-plugins

WORKDIR /app
ADD . /app

EXPOSE 8085 8087

CMD ["mvn", "spring-boot:run"]
