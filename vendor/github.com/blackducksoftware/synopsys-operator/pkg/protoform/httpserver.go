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

	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/gin-gonic/contrib/static"
	gin "github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
)

// SetupHTTPServer will used to create all the http api
func SetupHTTPServer(kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset, namespace string) {

	// all other http traffic
	go func() {
		// data, err := ioutil.ReadFile("/public/index.html")
		// Set the router as the default one shipped with Gin
		router := gin.Default()
		gin.SetMode(gin.ReleaseMode)
		// Serve frontend static files
		router.Use(static.Serve("/", static.LocalFile("/views", true)))

		// prints debug stuff out.
		router.Use(GinRequestLogger())

		// prometheus metrics
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())
		h := promhttp.Handler()
		router.GET("/metrics", func(c *gin.Context) {
			h.ServeHTTP(c.Writer, c.Request)
		})

		// router.GET("/sql-instances", func(c *gin.Context) {
		// 	// keys := []string{"pvc-000", "pvc-001", "pvc-002"}
		// 	keys, _ := util.ListHubPV(hubClient, namespace)
		// 	c.JSON(200, keys)
		// })

		// router.GET("/storage-classes", func(c *gin.Context) {
		// 	var storageList map[string]string
		// 	storageList = make(map[string]string)
		// 	storageClasses, err := util.ListStorageClass(kubeClient)
		// 	if err != nil {
		// 		log.Errorf("unable to list the storage classes due to %+v", err)
		// 		c.JSON(404, fmt.Sprintf("\"message\": \"Failed to List the storage class due to %+v\"", err))
		// 	}
		// 	for _, storageClass := range storageClasses.Items {
		// 		storageList[storageClass.GetName()] = fmt.Sprintf("%s (%s)", storageClass.GetName(), storageClass.Provisioner)
		// 	}
		// 	storageList["none"] = fmt.Sprintf("%s (%s)", "None", "Disable dynamic provisioner")
		// 	c.JSON(200, storageList)
		// })

		// router.GET("/hub", func(c *gin.Context) {
		// 	log.Debug("get hub request")
		// 	hubs, err := hubClient.SynopsysV1().Hubs(namespace).List(metav1.ListOptions{})
		// 	if err != nil {
		// 		log.Errorf("unable to get the hub list due to %+v", err)
		// 		c.JSON(404, "\"message\": \"Failed to List the hub\"")
		// 	}

		// 	log.Debugf("found %d hubs", len(hubs.Items))
		// 	returnVal := make(map[string]v1.Blackduck)

		// 	for _, v := range hubs.Items {
		// 		//l og.Debugf("hub %v: %+v", k, v)
		// 		returnVal[v.Spec.Namespace] = v
		// 	}
		// 	// log.Debugf("returnVal : %+v", returnVal)
		// 	c.JSON(200, returnVal)
		// })

		// router.POST("/hub", func(c *gin.Context) {
		// 	log.Debug("create hub request")
		// 	hubSpec := &v1.BlackduckSpec{}
		// 	if err := c.BindJSON(hubSpec); err != nil {
		// 		log.Debugf("Fatal failure binding the incoming request ! %v", c.Request)
		// 	}

		// 	ns, err := util.CreateNamespace(kubeClient, hubSpec.Namespace)
		// 	log.Debugf("created namespace: %+v", ns)
		// 	if err != nil {
		// 		log.Errorf("unable to create the namespace due to %+v", err)
		// 		c.JSON(404, "\"message\": \"Failed to create the namespace\"")
		// 		return
		// 	}
		// 	hubClient.SynopsysV1().Hubs(hubSpec.Namespace).Create(&v1.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: hubSpec.Namespace}, Spec: *hubSpec})

		// 	c.JSON(200, "\"message\": \"Succeeded\"")
		// })

		// router.DELETE("/hub", func(c *gin.Context) {
		// 	hubSpec := &v1.BlackduckSpec{}
		// 	if err := c.BindJSON(hubSpec); err != nil {
		// 		log.Debugf("Fatal failure binding the incoming request ! %v", c.Request)
		// 		return
		// 	}

		// 	log.Debugf("delete hub request %v", hubSpec.Namespace)

		// 	// This is on the event loop.
		// 	hubClient.SynopsysV1().Hubs(hubSpec.Namespace).Delete(hubSpec.Namespace, &metav1.DeleteOptions{})

		// 	c.JSON(200, "\"message\": \"Succeeded\"")
		// })

		// Start and run the server - blocking call, obviously :)
		router.Run(":8080")
	}()
}
