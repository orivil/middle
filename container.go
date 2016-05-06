// Copyright 2016 orivil Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middle provide a middleware container for store the middlewares and provide a
// names bag for config the middlewares.
package middle
import (
	"fmt"
	"gopkg.in/orivil/service.v0"
	"gopkg.in/orivil/sorter.v0"
)

type Container struct {
	public *service.Container
	bag *Bag
	priorities map[string]int
	cache map[string][]string
}

func NewContainer(bag *Bag, public *service.Container) *Container {
	c := &Container{
		bag: bag,
		priorities: make(map[string]int, 10),
		cache: make(map[string][]string, 20),
		public: public,
	}
	bag.SetMiddleChecker(c)
	return c
}

// Add for add a new middleware to the container
//
// name: the middleware name
// provider: could be a service provider, or just a function
// priority: this parameter is not necessary, default priority = 0, and biger
// priority comes out first
func (c *Container) Add(name string, provider interface{}, priority... int) {
	_priority := 0
	if len(priority) > 0 {
		_priority = priority[0]
	}
	var _provider func(sc *service.Container) interface{}
	if p, ok := provider.(func(sc *service.Container) interface{}); !ok {
		_provider = func(sc *service.Container) interface{} {
			return provider
		}
	} else {
		_provider = p
	}
	c.public.Add(name, _provider)
	c.priorities[name] = _priority
}

// GetMiddlesMsg for print the middleware message
func GetMiddlesMsg(c *Container, actions map[string]map[string]map[string]bool) (msg []string) {
	var mids []string
	var acts []string
	midMaxLen := 0
	for bundle, controllers := range actions {
		for controller, acs := range controllers {
			for a, _ := range acs {
				action := bundle + "." + controller + "." + a
				middles := c.Get(action)
				middlesStr := "[ "
				for _, middle := range middles {
					middlesStr += middle + "|"
				}
				if len(middles) == 0 {
					middlesStr += "nil|"
				}
				middlesStr = middlesStr[0:len(middlesStr)-1]
				middlesStr += " ]"
				mids = append(mids, middlesStr)
				acts = append(acts, action)
				if len(middlesStr) > midMaxLen {midMaxLen = len(middlesStr)}
			}
		}
	}
	space := "                                                                    "
	msg = make([]string, len(mids))
	for index, mid := range mids {
		mid += space[0: midMaxLen - len(mid)]
		msg[index] = mid + " => " + acts[index]
	}
	return msg
}

// Get get sorted middlewares
//
// action: full name like “package.controller.action”
func (c *Container) Get(action string) (middles []string) {
	if _middles, ok := c.cache[action]; ok {
		middles = _middles
	} else {
		middles = c.bag.GetMiddles(action)
		priorities := make(map[string]int, len(middles))
		for _, middle := range middles {
			priorities[middle] = c.priorities[middle]
		}
		sorter := sorter.NewPrioritySorter(priorities)
		middles = sorter.SortReverse()
		c.cache[action] = middles
	}
	return
}

// Get for get middleware instances form service container
func Get(action string, mc *Container, sc *service.Container) (middles []interface{}) {
	mids := mc.Get(action)
	middles = make([]interface{}, len(mids))
	for index, mid := range mids {
		middles[index] = sc.Get(mid)
	}
	return
}

// implement middle.MiddleChecker interface
func (c *Container) CheckExist(middleware string) error {
	if _, exist := c.priorities[middleware]; !exist {
		return fmt.Errorf("middle.Container: middleware %s not registerd", middleware)
	}
	return nil
}

