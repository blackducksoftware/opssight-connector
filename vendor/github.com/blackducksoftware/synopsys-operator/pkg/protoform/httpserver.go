/*
Copyright (C) 2018 Synopsys, Inc.

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

package protoform

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupHTTPServer will used to create all the http api
func SetupHTTPServer() {

	// all other http traffic
	go func() {
		// data, err := ioutil.ReadFile("/public/index.html")
		// Set the router as the default one shipped with Gin
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		// Serve frontend static files
		// router.Use(static.Serve("/", static.LocalFile("/views", true)))

		// prints debug stuff out.
		// router.Use(GinRequestLogger())

		// prometheus metrics
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())
		h := promhttp.Handler()
		router.GET("/metrics", func(c *gin.Context) {
			h.ServeHTTP(c.Writer, c.Request)
		})

		// Start and run the server - blocking call, obviously :)
		router.Run(":8080")
	}()
}
