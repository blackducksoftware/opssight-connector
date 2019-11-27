/*
Copyright (C) 2019 Synopsys, Inc.

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

package crdupdater

import (
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	log "github.com/sirupsen/logrus"
)

// Route stores the configuration to add or delete the route
type Route struct {
	config      *CommonConfig
	deployer    *util.DeployerHelper
	routeClient *routeclient.RouteV1Client
	routes      []*api.Route
	oldRoutes   map[string]routev1.Route
	newRoutes   map[string]*routev1.Route
}

// NewRoute returns the route configuration
func NewRoute(config *CommonConfig, routes []*api.Route) (*Route, error) {
	if !util.IsOpenshift(config.kubeClient) {
		return nil, nil
	}
	routeClient := util.GetRouteClient(config.kubeConfig, config.kubeClient, config.namespace)

	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newRoutes := append([]*api.Route{}, routes...)
	for i := 0; i < len(newRoutes); i++ {
		if !isLabelsExist(config.expectedLabels, newRoutes[i].Labels) {
			newRoutes = append(newRoutes[:i], newRoutes[i+1:]...)
			i--
		}
	}
	return &Route{
		config:      config,
		deployer:    deployer,
		routeClient: routeClient,
		routes:      newRoutes,
		oldRoutes:   make(map[string]routev1.Route, 0),
		newRoutes:   make(map[string]*routev1.Route, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new route
func (c *Route) buildNewAndOldObject() error {
	// build old route
	oldRoutes, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get routes for %s", c.config.namespace)
	}

	for _, oldRoute := range oldRoutes.(*routev1.RouteList).Items {
		c.oldRoutes[oldRoute.GetName()] = oldRoute
	}

	// build new route
	for i, newRoute := range c.routes {
		newRouteKube := util.GetRouteComponent(c.routeClient, newRoute, c.routes[i].Labels)
		c.newRoutes[newRoute.Name] = newRouteKube
	}
	return nil
}

// add adds the route
func (c *Route) add(isPatched bool) (bool, error) {
	for _, route := range c.routes {
		if _, ok := c.oldRoutes[route.Name]; !ok && !c.config.dryRun {
			log.Infof("creating Route %s", route.Name)
			_, err := util.CreateRoute(c.routeClient, c.config.namespace, c.newRoutes[route.Name])
			if err != nil {
				return false, errors.Annotatef(err, "unable to deploy route in %s", c.config.namespace)
			}
		}
	}
	return false, nil
}

// get gets the route
func (c *Route) get(name string) (interface{}, error) {
	return util.GetRoute(c.routeClient, c.config.namespace, name)
}

// list lists all the routes
func (c *Route) list() (interface{}, error) {
	return util.ListRoutes(c.routeClient, c.config.namespace, c.config.labelSelector)
}

// delete deletes the route
func (c *Route) delete(name string) error {
	log.Infof("deleting the route: %s", name)
	return util.DeleteRoute(c.routeClient, c.config.namespace, name)
}

// remove removes the route
func (c *Route) remove() error {
	// compare the old and new route and delete if needed
	for _, oldRoute := range c.oldRoutes {
		if _, ok := c.newRoutes[oldRoute.GetName()]; !ok {
			err := c.delete(oldRoute.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete route %s in namespace %s", oldRoute.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the route
func (c *Route) patch(cr interface{}, isPatched bool) (bool, error) {
	return false, nil
}
