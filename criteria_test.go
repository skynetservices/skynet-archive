package skynet

import (
	"testing"
)

type matchTestCase struct {
	Criteria             Criteria
	MatchingInstances    []ServiceInfo
	NonMatchingInstances []ServiceInfo
}

var testCases []matchTestCase = []matchTestCase{
	matchTestCase{
		Criteria: Criteria{
			Regions: []string{
				"Tampa",
			},
		},
		MatchingInstances: []ServiceInfo{
			ServiceInfo{
				ServiceConfig: &ServiceConfig{
					Region: "Tampa",
				},
			},
		},
		NonMatchingInstances: []ServiceInfo{
			ServiceInfo{
				ServiceConfig: &ServiceConfig{
					Region: "Chicago",
				},
			},
			ServiceInfo{
				ServiceConfig: &ServiceConfig{
					Region: "Dallas",
				},
			},
		},
	},
}

func TestMatch(t *testing.T) {
	for _, tc := range testCases {
		for _, i := range tc.MatchingInstances {
			if !tc.Criteria.Matches(i) {
				t.Fatal("Instance expected to match criteria and did not")
			}
		}

		for _, i := range tc.NonMatchingInstances {
			if tc.Criteria.Matches(i) {
				t.Fatal("Instance should not match criteria")
			}
		}
	}
}
