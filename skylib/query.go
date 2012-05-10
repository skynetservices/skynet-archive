package skylib

import (
  "log"
	"github.com/4ad/doozer"
	"encoding/json"
  "strings"
)

type Query struct {
  Service     string
  Version     string
  Host        string
  Region      string
  DoozerConn  *doozer.Conn
  DoozerRev   int64


  // Internal use only
  pathLength  int
  paths    map[string]*doozer.FileInfo
  files    map[string]*doozer.FileInfo
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

func (q *Query) search() {
  q.paths = make(map[string]*doozer.FileInfo, 0)
  q.files = make(map[string]*doozer.FileInfo, 0)

  if q.DoozerRev == 0 {
    q.DoozerRev = q.getCurrentDoozerRevision()
  }


  path := "/services"

  if q.Service != "" {
    path += "/" +q.Service

    if q.Version != "" {
      path += "/" +q.Version

      if q.Region != "" {
        path += "/" +q.Region

        if q.Host != "" {
          path += "/" +q.Host
        }
      }
    }
  }

  doozer.Walk(q.DoozerConn, q.DoozerRev, path, q, nil)
}

func (q *Query) FindHosts() (*[]string){
  q.pathLength = 6
  q.search()

  return q.matchingPaths()
}


func (q *Query) FindRegions() (*[]string){
  q.pathLength = 5
  q.search()

  return q.matchingPaths()
}

func (q *Query) FindServices() (*[]string){
  q.pathLength = 3
  q.search()

  return q.matchingPaths()
}

func (q *Query) FindServiceVersions() (*[]string){
  q.pathLength = 4
  q.search()

  return q.matchingPaths()
}

func (q *Query) FindInstances() (*[]*Service){
  q.search()

  results := make([]*Service, 0)

  // At this point we don't know which values were supplied 
  // if all prefixed dir's were supplied no filtering is needed, but this may be all nodes
  for path, _ := range q.files {
    var service Service

		data, _, err := q.DoozerConn.Get(path, &q.DoozerRev)
		if err != nil {
			log.Panic(err.Error())
		}

    err = json.Unmarshal(data, &service)

    if q.Service != "" && q.Service != service.Name {
      continue
    }

    if q.Version != "" && q.Version != service.Version {
      continue
    }

    if q.Region != "" && q.Region != service.Region {
      continue
    }

    if q.Host != "" && q.Host != service.IPAddress {
      continue
    }

    results = append(results, &service)
  }

  // make sure we can garbage collect
  // who knows how long the app might keep the query object around for
  q.paths = nil
  q.files = nil

  return &results
}


func (q *Query) matchingPaths() (*[]string){
  results := make([]string, 0)
  
  for path, dir := range q.paths {
    parts := strings.Split(path, "/")

    if !q.pathMatches(parts, path) {
      continue
    }

    results = append(results, dir.Name)
  }

  // reset internal variables also make sure we can garbage collect
  // who knows how long the app might keep the query object around for
  q.paths = nil
  q.files = nil
  q.pathLength = 0

  return &results
}

func (q *Query) pathMatches(parts []string, path string) (bool) {
    if len(parts) >= 3 && q.Service != "" && parts[2] != q.Service {
      return false
    }

    if  len(parts) >= 4 && q.Version != "" && parts[3] != q.Version {
      return false
    }

    if  len(parts) >= 5 && q.Region != "" && parts[4] != q.Region {
      return false
    }

    if  len(parts) >= 6 && q.Host != "" && parts[5] != q.Host {
      return false
    }

    return true
}

func (q *Query) getCurrentDoozerRevision() (int64){
	revision, err := q.DoozerConn.Rev()

	if err != nil {
		log.Panic(err.Error())
	}

  return revision
}
