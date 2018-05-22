package accesscontrol

import (
	"strings"
)

const (
	get     = 1 << 0
	head    = 1 << 1
	post    = 1 << 2
	put     = 1 << 3
	delete  = 1 << 4
	connect = 1 << 5
	options = 1 << 6
	trace   = 1 << 7
	patch   = 1 << 8
)

func bitFieldForMethod(method string) uint16 {
	switch method {
	case "GET":
		return get
	case "HEAD":
		return head
	case "POST":
		return post
	case "PUT":
		return put
	case "DELETE":
		return delete
	case "CONNECT":
		return connect
	case "OPTIONS":
		return options
	case "TRACE":
		return trace
	case "PATCH":
		return patch
	default:
		return 0
	}
}

// Node of the `ScopeRouteTree` - it has a component of the URL, the scope (if any) of the
// URI up to this node, supported methods on the URI, and children of this node (if any).
type node struct {
	urlComponent *string
	children     []node // FIXME: Make this nil'able (*[]node) to avoid allocating for every node.
	methods      uint16
	scope        *int // FIXME: Change this to array of scope indices and `methods` bit fields.
}

func (n *node) hasMethod(method string) bool {
	return n.methods&bitFieldForMethod(method) != 0
}

func newNode() node {
	return node{
		urlComponent: nil,
		children:     *new([]node),
		methods:      0,
		scope:        nil,
	}
}

// ScopeRouteTree helps with matching URLs against scope URIs rapidly.
// It associates routes based on the scope IDs.
//
// NOTE: This is slow when there are large number of URIs have same/matching base paths.
// For example, `/foo/*/baz`, `/foo/bar/*` and `/foo/*` are similar URIs and they
// all match against `/foo/bar/baz`.
type ScopeRouteTree struct {
	root node
}

// NewScopeRouteTree creates a new tree for managing scopes.
func NewScopeRouteTree() *ScopeRouteTree {
	return &ScopeRouteTree{
		root: newNode(),
	}
}

// AddRoute to this tree using the given method, URI and scope index.
func (t *ScopeRouteTree) AddRoute(method, uri string, scope int) {
	treeNode := &t.root
	cleanURI := strings.Split(strings.Trim(uri, "/"), "?")[0] // Trim "/" and remove query params.
	components := strings.Split(cleanURI, "/")

parent:
	for idx, component := range components {
		isFinalComponent := idx == len(components)-1
		component := strings.TrimSpace(component)
		isWildCard := strings.HasPrefix(component, ":") || component == "*" || component == ""

		for i, child := range treeNode.children {
			if (isWildCard && child.urlComponent == nil) || (child.urlComponent != nil && *child.urlComponent == component) {
				if child.scope == nil || (*child.scope == scope && isFinalComponent) {
					treeNode = &treeNode.children[i]
					continue parent
				}
			}
		}

		emptyNode := newNode()
		if !isWildCard {
			emptyNode.urlComponent = &component
		}

		treeNode.children = append(treeNode.children, emptyNode)
		treeNode = &treeNode.children[len(treeNode.children)-1]
	}

	treeNode.scope = &scope
	treeNode.methods |= bitFieldForMethod(method)
}

// GetMatchingScopes for the given method and URL.
func (t *ScopeRouteTree) GetMatchingScopes(method, url string) []int {
	scopes := *new([]int)
	cleanURL := strings.Split(strings.Trim(url, "/"), "?")[0] // Trim "/" and remove query params.
	components := strings.Split(cleanURL, "/")

	if len(components) == 0 {
		return scopes
	}

	candidates := []node{t.root}

	for _, component := range components {
		newCandidates := *new([]node)
		for _, candidate := range candidates {
			for _, child := range candidate.children {
				isChildWildCard := child.urlComponent == nil
				if isChildWildCard && child.hasMethod(method) {
					scopes = append(scopes, *child.scope)
				} else if isChildWildCard || *child.urlComponent == component {
					newCandidates = append(newCandidates, child)
				}
			}
		}

		candidates = newCandidates
	}

	for _, candidate := range candidates {
		if candidate.hasMethod(method) && candidate.scope != nil {
			scopes = append(scopes, *candidate.scope)
		}
	}

	return scopes
}
