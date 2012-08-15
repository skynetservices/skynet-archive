package client

import (
	"encoding/json"
	"fmt"
	"github.com/4ad/doozer"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"log"
	"strings"
)

type Query struct {
	Service    string
	Version    string
	Host       string
	Port       string
	Region     string
	Registered *bool
	DoozerConn *skynet.DoozerConnection
	DoozerRev  int64

	// Internal use only
	pathLength int
	paths      map[string]*doozer.FileInfo
	files      map[string]*doozer.FileInfo
}

func (q *Query) VisitDir(path string, f *doozer.FileInfo) bool {
	parts := strings.Split(path, "/")

	// If we know we are looking for dir's at a specified level no need to dig deeper
	if q.pathLength > 0 && len(parts) > q.pathLength {
		return false
	}

	// If we know the length we need for a proper path we don't need any leading paths
	if q.pathLength <= 0 || q.pathLength == len(parts) {
		q.paths[path] = f
	}

	return true
}

func (q *Query) VisitFile(path string, f *doozer.FileInfo) {
	q.files[path] = f
}

func (q *Query) makePath() (path string) {
	path = "/services"

	if q.Service == "" {
		return
	}
	path += "/" + q.Service

	if q.Version == "" {
		return
	}
	path += "/" + q.Version

	if q.Region == "" {
		return
	}
	path += "/" + q.Region

	if q.Host == "" {
		return
	}
	path += "/" + q.Host

	if q.Port == "" {
		return
	}
	path += "/" + q.Port

	return
}

func (q *Query) search() {
	q.paths = make(map[string]*doozer.FileInfo, 0)
	q.files = make(map[string]*doozer.FileInfo, 0)

	if q.DoozerRev == 0 {
		q.DoozerRev = q.getCurrentDoozerRevision()
	}

	path := q.makePath()

	q.DoozerConn.Walk(q.DoozerRev, path, q, nil)
}

func (q *Query) FindHosts() []string {
	q.pathLength = 6
	q.search()

	return q.matchingPaths()
}

func (q *Query) FindRegions() []string {
	q.pathLength = 5
	q.search()

	return q.matchingPaths()
}

func (q *Query) FindServices() []string {
	q.pathLength = 3
	q.search()

	return q.matchingPaths()
}

func (q *Query) FindServiceVersions() []string {
	q.pathLength = 4
	q.search()

	return q.matchingPaths()
}

func (q *Query) FindInstances() []*service.Service {
	q.search()

	results := make([]*service.Service, 0)

	// At this point we don't know which values were supplied 
	// if all prefixed dir's were supplied no filtering is needed, but this may be all nodes
	for path, _ := range q.files {
		var s service.Service

		data, _, err := q.DoozerConn.Get(path, q.DoozerRev)
		if err != nil {
			log.Panic(err.Error())
		}

		err = json.Unmarshal(data, &s)

		if q.Service != "" && q.Service != s.Config.Name {
			continue
		}

		if q.Version != "" && q.Version != s.Config.Version {
			continue
		}

		if q.Region != "" && q.Region != s.Config.Region {
			continue
		}

		if q.Host != "" && q.Host != s.Config.ServiceAddr.IPAddress {
			continue
		}

		if q.Port != "" && q.Port != fmt.Sprintf("%d", s.Config.ServiceAddr.Port) {
			continue
		}

		if q.Registered != nil && *q.Registered != s.Registered {
			continue
		}

		results = append(results, &s)
	}

	// make sure we can garbage collect
	// who knows how long the app might keep the query object around for
	q.paths = nil
	q.files = nil

	return results
}

func (q *Query) matchingPaths() []string {
	results := make([]string, 0)
	unique := make(map[string]string, 0)

	for path, dir := range q.paths {

		if !q.PathMatches(path) {
			continue
		}

		if _, ok := unique[dir.Name]; !ok {
			unique[dir.Name] = dir.Name
			results = append(results, dir.Name)
		}
	}

	// reset internal variables also make sure we can garbage collect
	// who knows how long the app might keep the query object around for
	q.paths = nil
	q.files = nil
	q.pathLength = 0

	return results
}

func (q *Query) PathMatches(path string) bool {
	parts := strings.Split(path, "/")

	if len(parts) >= 3 && q.Service != "" && parts[2] != q.Service {
		return false
	}

	if len(parts) >= 4 && q.Version != "" && parts[3] != q.Version {
		return false
	}

	if len(parts) >= 5 && q.Region != "" && parts[4] != q.Region {
		return false
	}

	if len(parts) >= 6 && q.Host != "" && parts[5] != q.Host {
		return false
	}

	if len(parts) >= 7 {
		fmt.Println(parts)
	}

	return true
}

func (q *Query) ServiceMatches(s service.Service) bool {
	if q.Service != "" && s.Config.Name != q.Service {
		return false
	}

	if q.Version != "" && s.Config.Version != q.Version {
		return false
	}

	if q.Region != "" && s.Config.Region != q.Region {
		return false
	}

	if q.Host != "" && s.Config.ServiceAddr.IPAddress != q.Host {
		return false
	}

	if q.Port != "" && fmt.Sprintf("%d", s.Config.ServiceAddr.Port) != q.Port {
		return false
	}

	if q.Registered != nil && s.Registered != *q.Registered {
		return false
	}

	return true
}

func (q *Query) Reset() {
	q.Service = ""
	q.Version = ""
	q.Region = ""
	q.Host = ""
	q.Registered = nil
	q.Port = ""
}

func (q *Query) getCurrentDoozerRevision() int64 {
	revision, err := q.DoozerConn.Rev()

	if err != nil {
		log.Panic(err.Error())
	}

	return revision
}
