package router

import (
	"regexp"
	"../request"
	"../response"
)
type NextFunc func(NextFunc)
type Middleware func(request *request.Request, response *response.Response, next func())

type Route struct{
	regex *regexp.Regexp
	handler Middleware
}

type Router struct {
	routes map[string][]*Route
}

func (r *Router) Init(){
	r.routes = make(map[string][]*Route)
	r.routes["get"] = []*Route{}
	r.routes["post"] = []*Route{}
	r.routes["put"] = []*Route{}
	r.routes["delete"] = []*Route{}
	r.routes["patch"] = []*Route{}
}

func (r* Router) addHandler(method string, url *regexp.Regexp, middleware Middleware){
	var route = &Route{}
	route.regex = url
	route.handler = middleware
	r.routes[method] = append(r.routes[method], route)
}

func (r* Router) Get(url string, middleware Middleware) *Router{
	r.addHandler("get", compileRegex(url), middleware)
	return r
}

func (r* Router) Post(url string, middleware Middleware) *Router{
	r.addHandler("post", compileRegex(url), middleware)
	return r
}

func (r* Router) Put(url string, middleware Middleware) *Router{
	r.addHandler("put", compileRegex(url), middleware)
	return r
}

func (r* Router) Patch(url string, middleware Middleware) *Router{
	r.addHandler("patch", compileRegex(url), middleware)
	return r
}

func (r* Router) Delete(url string, middleware Middleware) *Router{
	r.addHandler("delete", compileRegex(url), middleware)
	return r
}

func (r* Router) Use(middleware Middleware) *Router{
	var regex = compileRegex("(.*)")
	r.addHandler("get", regex, middleware)
	r.addHandler("post", regex, middleware)
	r.addHandler("put", regex, middleware)
	r.addHandler("patch", regex, middleware)
	r.addHandler("delete", regex, middleware)
	return r
}
func (r* Router) FindNext(index int, method string, url string, request *request.Request) (Middleware, int){
	var i = index
	for i < len(r.routes[method]){
		var route = r.routes[method][i]
		if route.regex.MatchString(url){
			var regex = route.regex.FindStringSubmatch(url)
			for i, name := range route.regex.SubexpNames() {
				if name != "" {
					request.Params[name] = regex[i]
				}
			}
			return route.handler, i
		}
		i++
	}
	return nil, -1
}

func compileRegex(url string) *regexp.Regexp {
	var i = 0
	var buffer = "/"
	var regexStr = "^"
	var endVariable = ">(?:[A-Za-z0-9\\-\\_\\$\\.\\+\\!\\*\\'\\(\\)\\,]+))"
	if url[0] == '/' {
		i++
	}
	for i < len(url) {
		if url[i] == '/' {
			// this is a new group parse the last part
			regexStr += buffer + "/"
			buffer = ""
			i++
		} else {
			if url[i] == ':' && ( (i-1 >=0 && url[i-1] == '/') || (i-1 == -1)) {
				// a variable found, lets read it
				var tempbuffer = "(?P<"
				var variableName = ""
				var variableNameDone = false
				var done = false
				var hasRegex = false
				var innerGroup = 0
				// lets branch in to look deeper
				i++
				for done != true && i < len(url) {
					if url[i] == '/' {
						if variableName != "" {
							if innerGroup == 0 {
								if hasRegex == false {
									tempbuffer += endVariable
								}
								done = true
								break
							}
						}
						tempbuffer = ""
						break;
					} else if url[i] == '(' {
						if variableNameDone == false {
							variableNameDone = true
							tempbuffer += ">"
							hasRegex = true
						}
						tempbuffer += string(url[i])
						if url[i - 1] != '\\' {
							innerGroup++
						}
					} else if url[i] == ')' {
						tempbuffer += string(url[i])
						if url[i - 1] != '\\' {
							innerGroup--
						}
					} else {
						if variableNameDone == false {
							variableName += string(url[i])
						}
						tempbuffer += string(url[i])
					}
					i++
				}
				if tempbuffer != "" {
					if hasRegex == false && done == false {
						tempbuffer += endVariable
					} else if hasRegex {
						tempbuffer += ")"
					}
					buffer += tempbuffer
				} else {
					panic("Invalid Route regex")
				}
			} else {
				buffer += string(url[i])
				i++
			}
		}
	}
	if buffer != "" {
		regexStr += buffer
	}
	return regexp.MustCompile(regexStr + "(?:[\\/]{0,1})$")
}