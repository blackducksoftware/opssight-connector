/*
Copyright (C) 2018 Black Duck Software, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package basicskyfire

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	dockerClient "github.com/fsouza/go-dockerclient"
)

type Docker struct {
	Client *dockerClient.Client
}

type Tag struct {
	Results []Result `json:"results"`
}

type Result struct {
	Name string `json:"name"`
}

type Image struct {
	imageName string
	version   string
	podName   string
}

var dockerRepos = []string{"centos", "alpine", "nginx", "busybox", "redis", "ubuntu", "mongo", "memcached", "mysql", "postgres",
	"node", "registry", "golang", "hello-world", "php", "mariadb", "elasticsearch", "docker", "wordpress", "rabbitmq",
	"haproxy", "ruby", "python", "openjdk", "logstash", "traefik", "debian", "tomcat", "influxdb", "java",
	"swarm", "jenkins", "kibana", "maven", "ghost", "nextcloud", "cassandra", "telegraf", "kong", "nats",
	"vault", "drupal", "fedora", "owncloud", "jruby", "sonarqube", "sentry", "solr", "gradle", "perl",
	"rethinkdb", "neo4j", "percona", "groovy", "amazonlinux", "rocket.chat", "buildpack-deps", "chronograf", "redmine", "jetty",
	"erlang", "couchdb", "pypy", "flink", "iojs", "couchbase", "zookeeper", "joomla", "django", "mono",
	"piwik", "eclipse-mosquitto", "ubuntu-debootstrap", "bash", "crate", "nats-streaming", "elixir", "arangodb", "kapacitor", "tomee",
	"haxe", "opensuse", "websphere-liberty", "adminer", "oraclelinux", "gcc", "orientdb", "rails", "mongo-express", "odoo",
	"neurodebian", "ros", "xwiki", "clojure", "irssi", "ibmjava", "aerospike", "notary", "rust", "composer", "backdrop", "swift", "php-zendserver",
	"r-base", "julia", "celery", "nuxeo", "docker-dev", "znc", "gazebo", "bonita", "cirros", "haskell", "plone", "hylang", "rapidoid", "geonetwork",
	"eggdrop", "storm", "rakudo-star", "convertigo", "spiped", "hello-seattle", "mediawiki", "ubuntu-upstart", "lightstreamer", "fsharp",
	"swipl", "glassfish", "thrift", "mageia", "photon", "open-liberty", "hola-mundo", "crux", "teamspeak", "sourcemage", "silverpeas",
	"hipache", "scratch", "clearlinux", "si", "euleros", "clefos", "matomo"}

func NewDocker() (cli *Docker, err error) {

	endpoint := "unix:///var/run/docker.sock"
	client, err := dockerClient.NewVersionedClient(endpoint, "1.24")

	return &Docker{
		Client: client,
	}, err
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func (d *Docker) GetDockerImages(imageCount int) []Image {
	var images []Image
	count := 0
	for _, dockerRepo := range dockerRepos {
		repos, _ := d.Client.SearchImages(dockerRepo)
		for _, repo := range repos {
			var body []byte
			var err error
			if strings.Contains(repo.Name, "/") {
				body, err = getHttpResponse(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/", repo.Name), 200)
			} else {
				body, err = getHttpResponse(fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags/", repo.Name), 200)
			}

			if err != nil {
				fmt.Errorf("Unable to get docker tags for the repo %s due to %+v", repo.Name, err)
			}

			var tags Tag
			err = json.Unmarshal(body, &tags)
			fmt.Errorf("Unable to unmarshall docker tag response for the repo %s due to %+v", repo.Name, err)

			podName := strings.Replace(repo.Name, "/", "-", -1)
			podName = strings.Replace(podName, ".", "-", -1)

			tagCount := 0
			randomCount := random(1, 10)
			fmt.Printf("Randome count: %d \n", randomCount)
			for _, tag := range tags.Results {
				images = append(images, Image{imageName: repo.Name, version: tag.Name, podName: fmt.Sprintf("%s%d", podName, tagCount)})
				count++
				if count == imageCount {
					return images
				}
				tagCount++
				if tagCount == randomCount {
					break
				}
			}
		}
	}
	return images
}
