package skynet

import (
	"encoding/json"
	"fmt"
	"github.com/skynetservices/doozer"
	"log"
	"path"
	"strings"
)

// Query is used for finding collections of service instances, according to the criteria it is given.
type Query struct {
	// Service is the name of the service being queried. Blank for all services.
	Service string
	// Version is the version of the service being queried. Blank for all versions.
	Version string
	// Host is the host of the service being queried. Blank for all hosts.
	Host string
	// Port is the port of the service being queried. Blank for all ports.
	Port string
	// Region is the region of the service being queried. Blank for all regions.
	Region string
	// UUID is the UUID of the service being queried. Blank for any UUID.
	UUID string

	// Registered is the registered status of the service. Nil for any status.
	Registered *bool

	DoozerConn *DoozerConnection
	doozerRev  int64

	// Internal use only
	pathLength int
	paths      map[string]*doozer.FileInfo
	files      map[string]*doozer.FileInfo
}

type queryVisitor Query

func (q *queryVisitor) VisitDir(path string, f *doozer.FileInfo) bool {
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

func (q *queryVisitor) VisitFile(path string, f *doozer.FileInfo) {
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

	q.doozerRev = q.getCurrentDoozerRevision()

	path := q.makePath()

	q.DoozerConn.Walk(q.doozerRev, path, (*queryVisitor)(q), nil)
}

// *Query.FindHosts() finds all the hosts with services that match the query.
func (q *Query) FindHosts() []string {
	q.pathLength = 6
	q.search()

	return q.matchingPaths()
}

// *Query.FindRegions() finds all the regions with services that match the query.
func (q *Query) FindRegions() []string {
	q.pathLength = 5
	q.search()

	return q.matchingPaths()
}

// *Query.FindServices() finds paths for all the services that match the query.
func (q *Query) FindServices() []string {
	q.pathLength = 3
	q.search()

	return q.matchingPaths()
}

// *Query.FindServiceVersions finds all the versions with services that match the query.
func (q *Query) FindServiceVersions() []string {
	q.pathLength = 4
	q.search()

	return q.matchingPaths()
}

// *Query.FindInstances fetches the services that match the query.
func (q *Query) FindInstances() []*ServiceInfo {
	q.search()

	results := make([]*ServiceInfo, 0)

	// At this point we don't know which values were supplied 
	// if all prefixed dir's were supplied no filtering is needed, but this may be all nodes
	for path, _ := range q.files {
		var s ServiceInfo

		data, _, err := q.DoozerConn.Get(path, q.doozerRev)
		if err != nil {
			log.Panic(err.Error())
		}

		err = json.Unmarshal(data, &s)

		if !q.ServiceMatches(s) {
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

	for p, dir := range q.paths {

		if !q.pathMatches(p) {
			continue
		}

		if _, ok := unique[dir.Name]; !ok {
			pathMatches := true

			// If Port or Registered supplied, we have to inspect files to ensure the path has a match in it
			if q.Port != "" || q.Registered != nil {
				pathMatches = false
				rev := q.DoozerConn.GetCurrentRevision()

				files, _ := q.DoozerConn.Getdirinfo(p, rev, 0, -1)

				if files != nil {
					for _, file := range files {
						data, _, err := q.DoozerConn.Get(path.Join(p, file.Name), rev)

						if err == nil {
							s := ServiceInfo{}
							err = json.Unmarshal(data, &s)

							if q.ServiceMatches(s) {
								pathMatches = true
								break
							}
						}
					}
				}
			}

			if pathMatches {
				unique[dir.Name] = dir.Name
				results = append(results, dir.Name)
			}
		}
	}

	// reset internal variables also make sure we can garbage collect
	// who knows how long the app might keep the query object around for
	q.paths = nil
	q.files = nil
	q.pathLength = 0

	return results
}

// We aren't able to match a path for a query on port or registered
func (q *Query) pathMatches(path string) bool {
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

	return true
}

// *Query.ServiceMatches indicates if the query matches the given service.
func (q *Query) ServiceMatches(s ServiceInfo) bool {
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

	if q.UUID != "" && s.Config.UUID != q.UUID {
		return false
	}

	return true
}

// *Query.Reset erases this query, making it match all services until fields are set.
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
