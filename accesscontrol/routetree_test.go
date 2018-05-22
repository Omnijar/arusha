package accesscontrol

import (
	"sort"
	"strings"
	"testing"
)

func TestRouteAddition(t *testing.T) {
	tree := NewScopeRouteTree()
	urls := [][]string{
		{"GET", "/foo/bar/baz"},
		{"POST", "/foo/bar"},
		{"PATCH", "/boo/booya"},
		{"DELETE", "/foo/*/bar"},
		{"PUT", "/boo/*"},
		{"OPTIONS", "/boo/:id"},
	}

	for i, route := range urls {
		tree.AddRoute(route[0], route[1], i+1000)
	}

	for i, route := range urls {
		node := tree.root
		components := strings.Split(route[1], "/")
		lastComponentIdx := len(components) - 1
		expectedScope := i + 1000

	outer:
		for i, component := range components {
			for _, child := range node.children {
				isWildCard := component == "*" || strings.HasPrefix(component, ":")
				if (child.urlComponent == nil && isWildCard) || (child.urlComponent != nil && *child.urlComponent == component) {
					node = child
					if i == lastComponentIdx && node.scope != nil && *node.scope == expectedScope {
						break outer
					}
				}
			}
		}

		if !node.hasMethod(route[0]) {
			t.Fatalf("expected path %s to have method %s", route[1], route[0])
		}

		if *node.scope-1000 != i {
			t.Fatalf("expected path %s to have scope %d", route[1], i+1000)
		}
	}
}

func TestRouteMatching(t *testing.T) {
	tree := NewScopeRouteTree()
	urls := [][]string{
		{"GET", "/foo/*"},
		{"GET", "/foo/bar/baz"},
		{"GET", "/foo/bar/*"},
		{"POST", "/foo/*/bar"},
		{"DELETE", "/foo/:boo/bar"},
		{"DELETE", "/foo/bar"},
		{"GET", "/booya/:id"},
		{"GET", "/booya"},
		{"POST", "/booya"},
		{"PUT", "/booya/:id"},
		{"DELETE", "/booya/:id"},
		{"GET", "/booya/foo/:id"},
		{"GET", "/booya/foo/"},
		{"POST", "/booya/foo/"},
		{"PUT", "/booya/foo/:id/"},
		{"DELETE", "/booya/foo/:id"},
	}

	for i, route := range urls {
		tree.AddRoute(route[0], route[1], i+1000)
	}

	// Test insertion of same route with same scopes. This shouldn't produce duplicate nodes.

	// FIXME: Enable this test, once the tree has been modified.

	// for i := 1; i <= 20; i++ {
	// 	tree.AddRoute("GET", urls[0][1], 1000)
	// 	tree.AddRoute("GET", urls[1][1], 1001)
	// 	tree.AddRoute("GET", urls[2][1], 1002)
	// 	tree.AddRoute("POST", urls[3][1], 1003)
	// 	tree.AddRoute("DELETE", urls[4][1], 1004)
	// }

	// Test different route with existing scope. This should merge with existing node.
	tree.AddRoute("HEAD", urls[0][1], 1000)
	tree.AddRoute("PUT", urls[3][1], 1003)

	tests := [][]string{
		{"HEAD", "/foo/bar"},
		{"HEAD", "/foo/bar/baz"},
		{"GET", "/foo/bar/baz"},
		{"GET", "/foo/bar/bleh"},
		{"POST", "/foo/booya/bar"},
		{"PUT", "/foo/booya/bar"},
		{"DELETE", "/foo/bar/bar"},
		{"DELETE", "/foo/bar"},
		{"GET", "/booya/abcdefg"},
		{"GET", "/booya/"},
		{"POST", "/booya/"},
		{"PUT", "/booya/abcdefg"},
		{"DELETE", "/booya/asdaffg"},
		{"PATCH", "/foo/bar"},
		{"GET", "/booya/foo/asdada"},
		{"GET", "/booya/foo/"},
		{"POST", "/booya/foo/"},
		{"PUT", "/booya/foo/asdasdadsa/"},
		{"DELETE", "/booya/foo/asdadassd"},
	}

	expectedScopes := [][]int{
		{1000},
		{1000},
		{1000, 1001, 1002},
		{1000, 1002},
		{1003},
		{1003},
		{1004},
		{1005},
		{1006},
		{1007},
		{1008},
		{1009},
		{1010},
		{},
		{1006, 1011},
		{1006, 1012},
		{1013},
		{1009, 1014},
		{1010, 1015},
	}

	for i := range tests {
		method, url, requiredScopes := tests[i][0], tests[i][1], expectedScopes[i]
		scopes := tree.GetMatchingScopes(method, url)
		sort.Ints(scopes)

		isEqual := true

		if len(scopes) == len(requiredScopes) {
			for i := range scopes {
				if scopes[i] != requiredScopes[i] {
					isEqual = false
					break
				}
			}
		} else {
			isEqual = false
		}

		if !isEqual {
			t.Fatalf("expected %s: %s to match scopes %v but found %v", method, url, requiredScopes, scopes)
		}
	}
}
